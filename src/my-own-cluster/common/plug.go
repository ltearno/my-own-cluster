package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type walker struct {
	db         *leveldb.DB
	it         iterator.Iterator
	basePrefix string
}

func (w *walker) Key() string {
	if w.it.Key() == nil {
		return ""
	}

	return string(w.it.Key()[len(w.basePrefix):])
}

func (w *walker) Seek(key string) bool {
	key = w.basePrefix + key

	if w.it.Key() != nil {
		// check that maybe we are already on it
		if strings.HasPrefix(string(w.it.Key()), key) {
			return true
		}
	}

	w.it.Seek([]byte(key))

	return strings.HasPrefix(string(w.it.Key()), key)
}

/**
URL plugging and routing
*/

func (o *Orchestrator) PlugFunction(method string, path string, name string, startFunction string) error {
	method = strings.ToLower(method)

	data := &PluggedFunction{
		Type:          "function",
		Name:          name,
		StartFunction: startFunction,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	o.db.Put([]byte(fmt.Sprintf("/plugs/byspec/%s/%s", method, path)), dataJSON, nil)

	fmt.Printf("plugged_function on method:%s, path:'%s', name:%s, start_function:%s\n", method, path, name, startFunction)

	return nil
}

func (o *Orchestrator) PlugFile(method string, path string, name string) error {
	method = strings.ToLower(method)

	// register plug
	data := &PluggedFile{
		Type: "file",
		Name: name,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	o.db.Put([]byte(fmt.Sprintf("/plugs/byspec/%s/%s", method, path)), dataJSON, nil)

	fmt.Printf("plugged_file on method:%s, path:'%s', name:%s\n", method, path, name)

	return nil
}

func (o *Orchestrator) UnplugPath(method string, path string) error {
	method = strings.ToLower(method)

	o.db.Delete([]byte(fmt.Sprintf("/plugs/byspec/%s/%s", method, path)), nil)

	fmt.Printf("unplugged_path on method:%s, path:'%s'\n", method, path)

	return nil
}

func (o *Orchestrator) GetPlugs() map[string]string {
	r := make(map[string]string, 0)

	prefix := []byte("/plugs/byspec/")

	iter := o.db.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()[len(prefix):]
		value := iter.Value()

		r[string(dup(key))] = string(dup(value))
	}
	iter.Release()

	return r
}

/****

BUG TO FIX :

when there is those path plugged :

/api/toto
/api/toto/!param

the query to /api/toto/coucou is 404 but should be routed to the second plug

*****/

func (o *Orchestrator) findPlug(method string, path string) (bool, string, interface{}, map[string]string) {
	walker := &walker{
		db:         o.db,
		it:         o.db.NewIterator(nil, nil),
		basePrefix: fmt.Sprintf("/plugs/byspec/%s/", method),
	}

	originalPath := path

	boundParameters := make(map[string]string)

	prefix := ""

	defer walker.it.Release()

	if !walker.Seek("") {
		fmt.Printf("no plugs registered (%s)\n", walker.Key())
		return false, "", nil, nil
	}

	// seek the beginning of the plug entries
	for len(path) > 0 {
		// path part that can be consummed (basically, everything until a '/')
		askedPathPart := path
		nextPartIndex := 1 + strings.Index(askedPathPart[1:], "/")
		if nextPartIndex > 0 {
			askedPathPart = askedPathPart[:nextPartIndex]
		}

		if o.trace {
			fmt.Printf("plug seek [%s] [%s] key:'%s' prefix:'%s' askedPathPart:'%s'\n", method, path, walker.Key(), prefix, askedPathPart)
		}

		starKey := prefix + "/!"
		if strings.HasPrefix(string(walker.Key()), starKey) {
			currentKey := walker.Key()
			partName := currentKey[len(prefix)+2:]
			nextPrefix := currentKey
			nextPartInKey := strings.Index(currentKey[len(prefix)+1:], "/")
			if nextPartInKey >= 0 {
				nextPrefix = currentKey[:nextPartInKey]
				partName = currentKey[len(prefix)+2 : nextPartInKey]
			}

			partValue := askedPathPart[1:]
			if len(partValue) == 0 {
				fmt.Printf("CANNOT HAVE AN EMPTY PATH PART VALUE\n")
				return false, "", nil, nil
			}

			prefix = nextPrefix
			path = path[len(askedPathPart):]

			// we have a plug that consumes the path part
			if o.trace {
				fmt.Printf("plug seek we have partName:'%s' = partValue:'%s'\n", partName, partValue)
			}
			boundParameters[partName] = partValue
		} else {
			prefix = prefix + askedPathPart
			path = path[len(askedPathPart):]

			ok := walker.Seek(prefix)
			if !ok {
				fmt.Printf("plug seek cannot seek to prefix:'%s'\n", prefix)
				return false, "", nil, nil
			}
		}
	}

	if o.trace {
		fmt.Printf("plug seek (%s) [%s] '%s' '%s' '%s'\n", method, path, walker.Key(), prefix, path)
	}

	if prefix != walker.Key() {
		fmt.Printf("PATH NOT MATCHING RESIDUAL KEY '%s'/'%s'\n", prefix, walker.Key())
		return false, "", nil, nil
	}

	if o.trace {
		fmt.Printf("plugged path '%s' matched with '%s'\n", walker.Key(), originalPath)
	}
	data := &Plug{}
	err := json.Unmarshal(walker.it.Value(), data)
	if err != nil {
		return false, "", nil, nil
	}

	switch data.Type {
	case "function":
		data := &PluggedFunction{}
		err = json.Unmarshal(walker.it.Value(), data)
		if err != nil {
			return false, "", nil, nil
		}
		return true, "function", data, boundParameters

	case "file":
		data := &PluggedFile{}
		err = json.Unmarshal(walker.it.Value(), data)
		if err != nil {
			return false, "", nil, nil
		}
		return true, "file", data, boundParameters
	}

	return false, "", nil, nil
}

func (o *Orchestrator) GetPlugFromPath(method string, path string) (bool, string, interface{}, map[string]string) {
	method = strings.ToLower(method)

	ok, plugType, plug, boundParameters := o.findPlug(method, path)

	return ok, plugType, plug, boundParameters
}
