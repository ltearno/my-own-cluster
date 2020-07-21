package common

import (
	"encoding/json"
	"fmt"

	"github.com/rs/xid"
)

var storageKey []byte = []byte(fmt.Sprintf("/filters"))

type Filter struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	StartFunction string `json:"start_function"`
	Data          string `json:"data"`
}

func (o *Orchestrator) PlugFilter(name string, startFunction string, data string) (string, error) {
	filters := make([]*Filter, 0)

	val, err := o.db.Get(storageKey, nil)
	if err == nil {
		json.Unmarshal(val, &filters)
	}

	filter := &Filter{
		ID:            xid.New().String(),
		Name:          name,
		StartFunction: startFunction,
		Data:          data,
	}

	filters = append(filters, filter)

	newVal, err := json.Marshal(filters)
	if err != nil {
		return "", err
	}

	o.db.Put(storageKey, newVal, nil)

	return filter.ID, nil
}

// Very unoptmized !
func (o *Orchestrator) UnplugFilter(id string) {
	filters := o.GetFilters()

	fmt.Printf("LKJHLKJH\n")

	r := make([]*Filter, 0)

	for _, f := range filters {
		if f.ID != id {
			r = append(r, &f)
		}
	}

	newVal, err := json.Marshal(r)
	if err == nil {
		o.db.Put(storageKey, newVal, nil)
	}

	fmt.Printf("LKJHLKJH %s\n", string(newVal))
}

func (o *Orchestrator) GetFilters() []Filter {
	res := make([]Filter, 0)
	val, err := o.db.Get(storageKey, nil)
	if err == nil {
		err = json.Unmarshal(val, &res)
	}

	return res
}
