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
	nextPortID  int32
	outputPorts map[int]*OutputPort

	lock sync.Mutex

	db *leveldb.DB
}

func NewOrchestrator(db *leveldb.DB) *Orchestrator {
	return &Orchestrator{
		nextPortID:  1,
		outputPorts: make(map[int]*OutputPort),
		db:          db,
	}
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

func (o *Orchestrator) CreateOutputPort() int {
	portID := atomic.AddInt32(&o.nextPortID, 1)
	o.outputPorts[int(portID)] = &OutputPort{
		buffer: []byte{},
	}
	return int(portID)
}

func (o *Orchestrator) GetOutputPort(portID int) *OutputPort {
	return o.outputPorts[portID]
}

func (o *Orchestrator) ReleaseOutputPort(portID int) {
	delete(o.outputPorts, portID)
}
