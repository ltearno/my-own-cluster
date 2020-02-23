package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

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

type FunctionExecutionContext struct {
	Orchestrator  *Orchestrator
	Name          string
	StartFunction string
	Trace         bool

	HasFinishedRunning     bool
	InputExchangeBufferID  int
	OutputExchangeBufferID int
	Result                 int
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
