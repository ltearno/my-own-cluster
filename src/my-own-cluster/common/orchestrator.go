package common

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
)

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

func (o *Orchestrator) PlugFunction(path string, name string, startFunction string) bool {
	o.db.Put([]byte(fmt.Sprintf("/function_plugs/bypath/%s/name", path)), []byte(name), nil)
	o.db.Put([]byte(fmt.Sprintf("/function_plugs/bypath/%s/start_function", path)), []byte(startFunction), nil)

	fmt.Printf("plugged_function '%s', name:%s, start_function:%s\n", path, name, startFunction)

	return true
}

func (o *Orchestrator) GetPluggedFunctionFromPath(path string) (string, string, bool) {
	name, err := o.db.Get([]byte(fmt.Sprintf("/function_plugs/bypath/%s/name", path)), nil)
	if err != nil {
		return "", "", false
	}

	startFunction, err := o.db.Get([]byte(fmt.Sprintf("/function_plugs/bypath/%s/start_function", path)), nil)
	if err != nil {
		return "", "", false
	}

	return string(name), string(startFunction), true
}

func (o *Orchestrator) RegisterFile(path string, contentType string, bytes []byte) string {
	techID := Sha256Sum(bytes)

	alreadyTechID, present := o.GetFileTechIDFromPath(path)
	if present && alreadyTechID == techID {
		return techID
	}

	o.db.Put([]byte(fmt.Sprintf("/files/byid/%s/content-type", techID)), []byte(contentType), nil)
	o.db.Put([]byte(fmt.Sprintf("/files/byid/%s/bytes", techID)), bytes, nil)
	o.db.Put([]byte(fmt.Sprintf("/files/bypath/%s", path)), []byte(techID), nil)

	fmt.Printf("registered_file '%s', size:%d, techID:%s\n", path, len(bytes), techID)

	return techID
}

func (o *Orchestrator) GetFileTechIDFromPath(path string) (string, bool) {
	techIDBytes, err := o.db.Get([]byte(fmt.Sprintf("/files/bypath/%s", path)), nil)
	if err != nil {
		return "", false
	}

	techID := string(techIDBytes)

	return techID, true
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
