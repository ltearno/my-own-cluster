package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var persistencePrefix = []byte("/persistence")

type RegisterFunction struct {
	TechId    string
	Name      string
	WasmBytes []byte
}

type Orchestrator struct {
	nextExchangeBufferID int32
	exchangeBuffers      map[int]*ExchangeBuffer

	lock sync.Mutex

	db *leveldb.DB
}

func NewOrchestrator(db *leveldb.DB) *Orchestrator {
	return &Orchestrator{
		nextExchangeBufferID: 0,
		exchangeBuffers:      make(map[int]*ExchangeBuffer),
		db:                   db,
	}
}

func (o *Orchestrator) PersistenceSet(key []byte, value []byte) bool {
	key = append(persistencePrefix, key...)

	o.db.Put(key, value, nil)

	return true
}

func (o *Orchestrator) PersistenceGet(key []byte) ([]byte, bool) {
	key = append(persistencePrefix, key...)

	value, err := o.db.Get(key, nil)
	if err != nil {
		return nil, false
	}

	return value, true
}

func dup(b []byte) []byte {
	r := make([]byte, len(b))
	copy(r, b)
	return r
}

func (o *Orchestrator) PersistenceGetSubset(keyPrefix []byte) ([][]byte, error) {
	keyPrefix = append(persistencePrefix, keyPrefix...)

	r := make([][]byte, 0)

	iter := o.db.NewIterator(util.BytesPrefix(keyPrefix), nil)
	for iter.Next() {
		key := iter.Key()[len(persistencePrefix):]
		value := iter.Value()

		r = append(r, dup(key), dup(value))
	}
	iter.Release()

	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return r, nil
}

/**
 * TODO
 *
 * We should be able to accept wildcards and named parameters in plugged path.
 *
 * The named parameters should then be injected as headers in the called input exchange buffer
 */

type Plug struct {
	Type string `json:"type"`
}

type PluggedFile struct {
	Type   string `json:"type"`
	TechID string `json:"tech_id"`
}

type PluggedFunction struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	StartFunction string `json:"start_function"`
}

func (o *Orchestrator) PlugFunction(method string, path string, name string, startFunction string) bool {
	method = strings.ToLower(method)

	data := &PluggedFunction{
		Type:          "function",
		Name:          name,
		StartFunction: startFunction,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return false
	}

	o.db.Put([]byte(fmt.Sprintf("/plugs/byspec/%s/%s", method, path)), dataJSON, nil)

	fmt.Printf("plugged_function on method:%s, path:'%s', name:%s, start_function:%s\n", method, path, name, startFunction)

	return true
}

func (o *Orchestrator) PlugFile(method string, path string, contentType string, bytes []byte) string {
	method = strings.ToLower(method)

	// register content
	techID := Sha256Sum(bytes)
	o.db.Put([]byte(fmt.Sprintf("/files/byid/%s/content-type", techID)), []byte(contentType), nil)
	o.db.Put([]byte(fmt.Sprintf("/files/byid/%s/bytes", techID)), bytes, nil)

	// register plug
	data := &PluggedFile{
		Type:   "file",
		TechID: techID,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	o.db.Put([]byte(fmt.Sprintf("/plugs/byspec/%s/%s", method, path)), dataJSON, nil)

	fmt.Printf("registered_file '%s', size:%d, techID:%s\n", path, len(bytes), techID)

	return techID
}

func getString(db *leveldb.DB, key string) (string, bool) {
	bytes, err := db.Get([]byte(key), nil)
	if err != nil {
		return "", false
	}

	return string(bytes), true
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

	//fmt.Printf("current key: '%s'\n", string(w.it.Key()))

	return strings.HasPrefix(string(w.it.Key()), key)
}

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

		//fmt.Printf("(%s) [%s] '%s' '%s' '%s', currently_matching:'%s'\n", method, path, walker.Key(), prefix, path, askedPathPart)

		starKey := prefix + "/!"
		if strings.HasPrefix(string(walker.Key()), starKey) {
			//ok := walker.Seek(starKey)
			//if ok {
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
			fmt.Printf(" we have '%s' = '%s'\n", partName, partValue)
			boundParameters[partName] = partValue
		} else {
			prefix = prefix + askedPathPart
			path = path[len(askedPathPart):]

			ok := walker.Seek(prefix)
			if !ok {
				fmt.Printf("PLUG NOT FOUND (prefix:%s)\n", prefix)
				return false, "", nil, nil
			}
		}
	}

	fmt.Printf("(%s) [%s] '%s' '%s' '%s'\n", method, path, walker.Key(), prefix, path)

	if prefix != walker.Key() {
		fmt.Printf("PATH NOT MATCHING RESIDUAL KEY '%s'/'%s'\n", prefix, walker.Key())
		return false, "", nil, nil
	}

	fmt.Printf("PATH FULLY MATCHED, WE HAVE A PLUG for path '%s' mathed to '%s'\n", originalPath, walker.Key())
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

func (o *Orchestrator) GetFileContentType(techID string) (string, bool) {
	contentTypeBytes, err := o.db.Get([]byte(fmt.Sprintf("/files/byid/%s/content-type", techID)), nil)
	if err != nil {
		return "", false
	}

	contentType := string(contentTypeBytes)

	return contentType, true
}

func (o *Orchestrator) GetFileBytes(techID string) ([]byte, bool) {
	fileBytes, err := o.db.Get([]byte(fmt.Sprintf("/files/byid/%s/bytes", techID)), nil)
	if err != nil {
		return nil, false
	}

	return fileBytes, true
}

func (o *Orchestrator) RegisterFunction(name string, wasmBytes []byte) string {
	techID := Sha256Sum(wasmBytes)

	alreadyTechID, present := o.GetFunctionTechIDFromName(name)
	if present && alreadyTechID == techID {
		return techID
	}

	o.db.Put([]byte(fmt.Sprintf("/functions/byid/%s/bytes", techID)), wasmBytes, nil)
	o.db.Put([]byte(fmt.Sprintf("/functions/byname/%s", name)), []byte(techID), nil)

	fmt.Printf("registered_function '%s', size:%d, techID:%s\n", name, len(wasmBytes), techID)

	return techID
}

func (o *Orchestrator) GetFunctionTechIDFromName(name string) (string, bool) {
	techIDBytes, err := o.db.Get([]byte(fmt.Sprintf("/functions/byname/%s", name)), nil)
	if err != nil {
		return "", false
	}

	techID := string(techIDBytes)

	return techID, true
}

func (o *Orchestrator) GetFunctionBytes(techID string) ([]byte, bool) {
	wasmBytes, err := o.db.Get([]byte(fmt.Sprintf("/functions/byid/%s/bytes", techID)), nil)
	if err != nil {
		return nil, false
	}

	return wasmBytes, true
}

func (o *Orchestrator) GetFunctionBytesByFunctionName(functionName string) ([]byte, bool) {
	techID, ok := o.GetFunctionTechIDFromName(functionName)
	if !ok {
		return nil, false
	}

	wasmBytes, ok := o.GetFunctionBytes(techID)
	if !ok {
		return nil, false
	}

	return wasmBytes, true
}

func (o *Orchestrator) CreateExchangeBuffer() int {
	bufferID := atomic.AddInt32(&o.nextExchangeBufferID, 1)
	o.exchangeBuffers[int(bufferID)] = newExchangeBuffer()
	return int(bufferID)
}

func (o *Orchestrator) GetExchangeBuffer(bufferID int) *ExchangeBuffer {
	return o.exchangeBuffers[bufferID]
}

func (o *Orchestrator) ReleaseExchangeBuffer(bufferID int) {
	delete(o.exchangeBuffers, bufferID)
}
