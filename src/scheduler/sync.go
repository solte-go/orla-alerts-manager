package scheduler

import (
	"sync"
)

type syncContainer struct {
	tasks map[string]*taskWrapper
	mx    sync.Mutex
}

func newSyncContainer() syncContainer {
	return syncContainer{
		tasks: make(map[string]*taskWrapper),
	}
}

func (c *syncContainer) access(sync func(map[string]*taskWrapper)) {
	c.mx.Lock()
	defer c.mx.Unlock()

	sync(c.tasks)
}
