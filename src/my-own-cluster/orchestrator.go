package main

import (
	"fmt"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

type RegisterFunction struct {
	TechId    string
	Name      string
	WasmBytes []byte
}

type Orchestrator struct {
	nextPortID  int
	outputPorts map[int]*OutputPort

	lock sync.Mutex

	registeredFunctionsByName   map[string]string
	registeredFunctionsByTechId map[string]RegisterFunction

	db *leveldb.DB
}

func NewOrchestrator(db *leveldb.DB) *Orchestrator {
	return &Orchestrator{
		nextPortID:                  1,
		outputPorts:                 make(map[int]*OutputPort),
		registeredFunctionsByName:   make(map[string]string),
		registeredFunctionsByTechId: make(map[string]RegisterFunction),
		db:                          db,
	}
}

func (o *Orchestrator) RegisterFile(path string, contentType string, bytes []byte) string {
	techID := Sha256Sum(bytes)

	o.db.Put([]byte(fmt.Sprintf("/files/byid/%s/content-type", techID)), []byte(contentType), nil)
	o.db.Put([]byte(fmt.Sprintf("/files/byid/%s/bytes", techID)), bytes, nil)
	o.db.Put([]byte(fmt.Sprintf("/files/bypath/%s", path)), []byte(techID), nil)

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

func (o *Orchestrator) HasFunction(name string) bool {
	_, ok := o.registeredFunctionsByName[name]
	return ok
}

func (o *Orchestrator) RegisterFunction(name string, wasmBytes []byte) string {
	techID := Sha256Sum(wasmBytes)

	function := RegisterFunction{
		TechId:    techID,
		Name:      name,
		WasmBytes: wasmBytes,
	}
	o.registeredFunctionsByTechId[techID] = function

	alreadyRegisteredTechID, alreadyRegistered := o.registeredFunctionsByName[name]
	if alreadyRegistered && alreadyRegisteredTechID == techID {
		return techID
	}

	fmt.Printf("registered_function '%s', size:%d, techID:%s\n", name, len(wasmBytes), techID)

	o.registeredFunctionsByName[name] = techID

	return techID
}

func (o *Orchestrator) CreateOutputPort() int {
	o.lock.Lock()
	portID := o.nextPortID
	o.outputPorts[portID] = &OutputPort{
		closed:    false,
		buffer:    []byte{},
		listeners: []*chan []byte{},
	}
	o.nextPortID++
	o.lock.Unlock()
	return portID
}

func appendSlice(slice []byte, data []byte) []byte {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]byte, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func (o *Orchestrator) GetOutputPort(portID int) *OutputPort {
	return o.outputPorts[portID]
}

type OutputPort struct {
	closed    bool
	buffer    []byte
	listeners []*chan []byte
}

func (p *OutputPort) RegisterChannel() chan []byte {
	c := make(chan []byte, 1)
	p.listeners = append(p.listeners, &c)
	return c
}

func (p *OutputPort) GetBuffer() []byte {
	return p.buffer
}

func (p *OutputPort) Read(buffer []byte) int {
	return 0
}

func (p *OutputPort) Write(buffer []byte) int {
	// TODO WARNING THIS DOES NOT TAKE WRITE POS IN ACCOUNT !!!
	p.buffer = appendSlice(p.buffer, buffer)

	return len(buffer)
}

func (p *OutputPort) Close() int {
	p.closed = true

	for _, listener := range p.listeners {
		*listener <- p.buffer
	}

	return 0
}
