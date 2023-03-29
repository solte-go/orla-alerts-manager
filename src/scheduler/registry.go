package scheduler

import (
	"fmt"
	"orla-alerts/solte.lab/src/config"
	"orla-alerts/solte.lab/src/toolbox/db"
	"time"
)

var Registry = map[string]func(duration time.Duration, db *db.DB, tasks *config.Tasks) (Runnable, error){}

// TODO
func Add(name string, constructor func(duration time.Duration, store *db.DB, tasks *config.Tasks) (Runnable, error)) error {
	if _, registered := Registry[name]; registered {
		return fmt.Errorf("task %s already registered", name)
	}
	Registry[name] = constructor

	return nil
}
