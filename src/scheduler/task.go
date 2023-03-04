package scheduler

import (
	"time"
)

type taskStatus int

const (
	statusNotRunning taskStatus = iota
	statusRunning
)

func (t taskStatus) String() string {
	switch t {
	case statusNotRunning:
		return "not running"
	case statusRunning:
		return "running"
	}

	return "unknown status"
}

type Runnable interface {
	Run() error
}

type taskWrapper struct {
	Runnable
	status        taskStatus
	duration      time.Duration
	nextRun       time.Time
	lastRun       time.Time
	lastError     error
	runOnNextTick bool
}

type TaskInfo struct {
	Status    taskStatus    `json:"-"`
	StatusStr string        `json:"status"`
	Duration  time.Duration `json:"interval"`
	NextRun   time.Time     `json:"nextRun"`
	LastRun   time.Time     `json:"lastRun"`
	LastError error         `json:"lastError"`
}

func (tw taskWrapper) Ready() bool {
	return (tw.status != statusRunning && tw.deadlineReached()) || tw.runOnNextTick && tw.status != statusRunning
}

func (tw taskWrapper) deadlineReached() bool {
	if tw.nextRun.IsZero() {
		return false
	}

	return time.Now().After(tw.nextRun)
}

func (tw taskWrapper) TaskInfo() TaskInfo {
	return TaskInfo{
		Status:    tw.status,
		StatusStr: tw.status.String(),
		Duration:  time.Duration(tw.duration.Minutes()),
		NextRun:   tw.nextRun,
		LastRun:   tw.lastRun,
		LastError: tw.lastError,
	}
}

func (tw *taskWrapper) ScheduleNextRun() {
	if tw.duration != 0 {
		if time.Now().After(tw.nextRun) { //This is required for correct execution of Run-Once functionality.
			tw.nextRun = tw.nextRun.Add(tw.duration)
		}
	}
	tw.runOnNextTick = false

}
