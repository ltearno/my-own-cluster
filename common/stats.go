package common

type StatName string

var STAT_NB_CREATED_BUFFERS StatName = "nb_created_buffers"
var STAT_NB_REQUESTS_RECEIVED StatName = "nb_received_request"
var STAT_NB_CURRENT_BUFFERS StatName = "nb_current_buffers"

func (o *Orchestrator) StatIncrement(name StatName) {
	o.statsLock.Lock()

	actual, ok := o.stats[string(name)]
	if ok {
		o.stats[string(name)] = actual + 1
	} else {
		o.stats[string(name)] = 1
	}

	o.statsLock.Unlock()
}

func (o *Orchestrator) StatSet(name StatName, value int) {
	o.statsLock.Lock()
	o.stats[string(name)] = value
	o.statsLock.Unlock()
}
