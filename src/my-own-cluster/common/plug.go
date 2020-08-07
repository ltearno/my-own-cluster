package common

import (
	"fmt"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/****

BUG TO FIX :

when there is those path plugged :

/api/toto
/api/toto/!param

the query to /api/toto/coucou is 404 but should be routed to the second plug

*****/

type PlugSystem struct {
	db         *leveldb.DB
	identifier string
	trace      bool
}

func NewPlugSystem(db *leveldb.DB, identifier string, trace bool) *PlugSystem {
	return &PlugSystem{
		db:         db,
		identifier: identifier,
		trace:      trace,
	}
}

func (p *PlugSystem) getPlugsStartKey() []byte {
	return []byte(fmt.Sprintf("/plug_system/%s/byspec/", p.identifier))
}

func (p *PlugSystem) getPlugsStartKeyByMethod(method string) []byte {
	return []byte(fmt.Sprintf("/plug_system/%s/byspec/%s/", p.identifier, method))
}

func (p *PlugSystem) getPlugKey(method string, path string) []byte {
	return []byte(fmt.Sprintf("/plug_system/%s/byspec/%s/%s", p.identifier, method, path))
}

/**
URL plugging and routing
*/

func (p *PlugSystem) PlugPath(method string, path string, data []byte) error {
	method = strings.ToLower(method)

	p.db.Put(p.getPlugKey(method, path), data, &opt.WriteOptions{Sync: true})

	return nil
}

func (p *PlugSystem) UnplugPath(method string, path string) error {
	method = strings.ToLower(method)

	p.db.Delete(p.getPlugKey(method, path), nil)

	fmt.Printf("unplugged_path '%s' on method:%s, path:'%s'\n", p.identifier, method, path)

	return nil
}

func (p *PlugSystem) GetPlugs() map[string]string {
	r := make(map[string]string, 0)

	prefix := p.getPlugsStartKey()

	iter := p.db.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()[len(prefix):]
		value := iter.Value()

		r[string(dup(key))] = string(dup(value))
	}
	iter.Release()

	return r
}

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

func (p *PlugSystem) findPlug(method string, path string) (bool, []byte, map[string]string) {
	if p.trace {
		fmt.Printf("START findPlug '%s' '%s'\n", method, path)
	}

	walker := &walker{
		db:         p.db,
		it:         p.db.NewIterator(nil, nil),
		basePrefix: string(p.getPlugsStartKeyByMethod(method)),
	}

	originalPath := path

	boundParameters := make(map[string]string)

	prefix := ""

	defer walker.it.Release()

	if !walker.Seek("") {
		fmt.Printf("no plugs registered (%s)\n", walker.Key())
		return false, nil, nil
	}

	// seek the beginning of the plug entries
	for len(path) > 0 {
		// path part that can be consummed (basically, everything until a '/')
		askedPathPart := path
		nextPartIndex := 1 + strings.Index(askedPathPart[1:], "/")
		if nextPartIndex > 0 {
			askedPathPart = askedPathPart[:nextPartIndex]
		}

		if p.trace {
			fmt.Printf("plug '%s' seek [%s] [%s] key:'%s' prefix:'%s' askedPathPart:'%s'\n", p.identifier, method, path, walker.Key(), prefix, askedPathPart)
		}

		if strings.HasPrefix(string(walker.Key()), prefix+"/*") {
			currentKey := walker.Key()
			partName := currentKey[len(prefix)+2:]

			// we have a plug that consumes the whole path
			if p.trace {
				fmt.Printf("plug '%s' seek we have partName:'%s' = partValue:'%s'\n", p.identifier, partName, path[1:])
			}
			boundParameters[partName] = path[1:]

			// fucked up surely
			prefix = walker.Key()
			path = ""
		} else if strings.HasPrefix(string(walker.Key()), prefix+"/!") {
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
				return false, nil, nil
			}

			prefix = nextPrefix
			path = path[len(askedPathPart):]

			// we have a plug that consumes the path part
			if p.trace {
				fmt.Printf("plug '%s' seek we have partName:'%s' = partValue:'%s'\n", p.identifier, partName, partValue)
			}
			boundParameters[partName] = partValue
		} else { // general case
			prefix = prefix + askedPathPart
			path = path[len(askedPathPart):]

			ok := walker.Seek(prefix)
			if !ok {
				fmt.Printf("plug '%s' seek cannot seek to prefix:'%s'\n", p.identifier, prefix)
				return false, nil, nil
			}
		}
	}

	if p.trace {
		fmt.Printf("plug '%s' seek (%s) [%s] '%s' '%s' '%s'\n", p.identifier, method, path, walker.Key(), prefix, path)
	}

	if prefix != walker.Key() {
		fmt.Printf("'%s' PATH NOT MATCHING RESIDUAL KEY '%s'/'%s'\n", p.identifier, prefix, walker.Key())
		return false, nil, nil
	}

	if p.trace {
		fmt.Printf("plugged '%s' path '%s' matched with '%s'\n", p.identifier, walker.Key(), originalPath)
	}

	return true, walker.it.Value(), boundParameters
}
