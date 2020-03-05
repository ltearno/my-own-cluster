package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type ExecutionEngineContext interface {
	Run() error
}

type ExecutionEngine interface {
	PrepareContext(fctx *FunctionExecutionContext) (ExecutionEngineContext, error)
}

type ExecutionEngineContextBounding interface{}

type APIProvider interface {
	BindToExecutionEngineContext(ctx ExecutionEngineContextBounding)
}

type Orchestrator struct {
	nextExchangeBufferID int32
	exchangeBuffers      map[int]*ExchangeBuffer

	lock sync.Mutex

	db *leveldb.DB

	executionEngines map[string]ExecutionEngine
	apiProviders     map[string]APIProvider
}

func NewOrchestrator(db *leveldb.DB) *Orchestrator {
	return &Orchestrator{
		nextExchangeBufferID: 0,
		exchangeBuffers:      make(map[int]*ExchangeBuffer),
		db:                   db,
		executionEngines:     make(map[string]ExecutionEngine),
		apiProviders:         make(map[string]APIProvider),
	}
}

func (o *Orchestrator) AddExecutionEngine(contentType string, engine ExecutionEngine) {
	o.executionEngines[contentType] = engine
	fmt.Printf("added execution engine for '%s'\n", contentType)
}

func (o *Orchestrator) AddAPIProvider(moduleName string, apiProvider APIProvider) {
	o.apiProviders[moduleName] = apiProvider
	fmt.Printf("added api provider for '%s'\n", moduleName)
}

func (o *Orchestrator) GetAPIProvider(moduleName string) APIProvider {
	v, present := o.apiProviders[moduleName]
	if !present {
		return nil
	}

	return v
}

type FunctionExecutionContext struct {
	Orchestrator *Orchestrator

	CodeBytes     []byte
	Name          string
	StartFunction string
	Arguments     []int

	Trace                 bool
	Mode                  string // direct or posix
	POSIXFileName         *string
	POSIXArguments        *[]string
	InputExchangeBufferID int

	HasFinishedRunning     bool
	OutputExchangeBufferID int
	Result                 int
}

/*

Execution context creation and launch

*/

func (o *Orchestrator) NewFunctionExecutionContext(
	functionName string,
	startFunction string,
	arguments []int,
	trace bool,
	mode string,
	posixFileName *string,
	posixArguments *[]string,
	inputExchangeBufferID int,
	outputExchangeBufferID int,
) *FunctionExecutionContext {
	return &FunctionExecutionContext{
		Orchestrator: o,

		Name:          functionName,
		StartFunction: startFunction,

		Trace:          trace,
		Mode:           mode,
		Arguments:      arguments,
		POSIXFileName:  posixFileName,
		POSIXArguments: posixArguments,

		HasFinishedRunning:     false,
		InputExchangeBufferID:  inputExchangeBufferID,
		OutputExchangeBufferID: outputExchangeBufferID,
		Result:                 -1,
	}
}

func (fctx *FunctionExecutionContext) Run() error {
	pluggedFunctionTechID, err := fctx.Orchestrator.GetBlobTechIDFromReference(fctx.Name)
	if err != nil {
		return fmt.Errorf("can't find plugged function (%s)", fctx.Name)
	}

	pluggedFunctionAbstract, err := fctx.Orchestrator.GetBlobAbstractByTechID(pluggedFunctionTechID)
	if err != nil {
		return fmt.Errorf("can't find plugged function abstract (%s)", fctx.Name)
	}

	codeBytes, err := fctx.Orchestrator.GetBlobBytesByTechID(pluggedFunctionTechID)
	if err != nil {
		return fmt.Errorf("can't find plugged function bytes (%s)", fctx.Name)
	}

	fctx.CodeBytes = codeBytes

	engine, hasEngine := fctx.Orchestrator.executionEngines[pluggedFunctionAbstract.ContentType]
	if hasEngine {
		ectx, err := engine.PrepareContext(fctx)
		if err != nil || ectx == nil {
			return fmt.Errorf("cannot create %s context for function: %v", pluggedFunctionAbstract.ContentType, err)
		}

		err = ectx.Run()
		if err != nil {
			return fmt.Errorf("execution %s error in function: %v", pluggedFunctionAbstract.ContentType, err)
		}
	} else {
		return fmt.Errorf("unknown function type '%s'", pluggedFunctionAbstract.ContentType)
	}

	if fctx.Trace {
		fmt.Printf(" -> result:%d\n", fctx.Result)
	}

	/*
		Instead of waiting for the end of the call, we should count references to the exchange buffer
		and wait for the last reference to dissappear. At this moment, the http response is complete and
		can be sent back to the client. This allows the first callee to transfer its output exchange
		buffer to another function and exit. The other function will then do whatever it wants to do
		(fan out, fan in and so on...).
		At this point in time, resources associated with this call can be destroyed.
	*/

	return nil
}

/*

Persistence service for applications, a simple key value database

*/

var persistencePrefix = []byte("/persistence")

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

type Plug struct {
	Type string `json:"type"`
}

type PluggedFile struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type PluggedFunction struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	StartFunction string `json:"start_function"`
}

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

type BlobNameStatus struct {
	Name   string `json:"name"`
	TechID string `json:"tech_id"`
}

func (o *Orchestrator) GetBlobsByName() []BlobNameStatus {
	r := make([]BlobNameStatus, 0)

	prefix := []byte("/blobs/byname/")

	iter := o.db.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()[len(prefix):]
		value := iter.Value()

		r = append(r, BlobNameStatus{
			Name:   string(dup(key)),
			TechID: string(dup(value)),
		})
	}
	iter.Release()

	return r
}

type BlobStatus struct {
	TechID      string `json:"tech_id"`
	ContentType string `json:"content_type"`
	Length      int    `json:"length"`
}

func (o *Orchestrator) GetBlobs() []BlobStatus {
	r := make([]BlobStatus, 0)

	prefix := []byte("/blobs/abstract/")

	iter := o.db.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := iter.Key()[len(prefix):]
		value := iter.Value()

		a := &BlobAbstract{}
		err := json.Unmarshal(value, a)
		if err != nil {
			continue
		}

		r = append(r, BlobStatus{
			TechID:      string(dup(key)),
			ContentType: a.ContentType,
			Length:      a.Length,
		})
	}
	iter.Release()

	return r
}

/**
GetStatus

*/

type status struct {
	NbExchangeBuffers  int               `json:"nb_exchange_buffers"`
	NbExchangedBuffers int               `json:"nb_exchanged_buffers"`
	Plugs              map[string]string `json:"plugs"`
	BlobNames          []BlobNameStatus  `json:"blob_names"`
	Blobs              []BlobStatus      `json:"blobs"`
}

func (o *Orchestrator) GetStatus() string {
	status := &status{}

	status.NbExchangeBuffers = len(o.exchangeBuffers)
	status.NbExchangedBuffers = int(o.nextExchangeBufferID)
	status.Plugs = o.GetPlugs()
	status.BlobNames = o.GetBlobsByName()
	status.Blobs = o.GetBlobs()

	b, err := json.Marshal(status)
	if err != nil {
		return ""
	}

	return string(b)
}
