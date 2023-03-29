package worker

import (
	"orla-alerts/solte.lab/src/model"
	"time"
)

type taskStatus int

const (
	notRunning taskStatus = iota
	running
	completedWithError
)

type job struct {
	status           taskStatus
	alerts           []*model.Alert   //alerts struct
	deliveryChan     chan model.Alert // channel for delivery from RabbitMQ
	bunchSize        int              //maximum number of domain data per loop
	currentBunchSize int              //current bunch size
	startTime        time.Time        //time when task was started
	timeout          time.Duration    //timeout for the task
	connRetryDelay   time.Duration    //task global retry for all functionsq
}

func (j *job) resetBunch() {
	j.currentBunchSize = 0
	j.alerts = j.alerts[:0]
}

func (j *job) addToBunch(record model.Alert) {
	j.currentBunchSize++
	j.alerts = append(j.alerts, &record)
}

func (j *job) ready() bool {
	if j.status != running {
		return true
	}
	return false
}

func (j *job) running() {
	j.status = running
}

func (j *job) notRunning() {
	j.status = notRunning
}

func (j *job) returnedWithError() {
	j.status = completedWithError
}

func (j *job) completedSuccessfully() {
	j.status = notRunning
}

// func (t *job) appliesMaxRequestDelay() {
// 	time.Sleep(t.maxActiveRequestsDelay)
// 	t.status = notRunning
// }
