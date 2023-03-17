package model

import (
	"math/rand"
	"time"
)

func GetTestAlert() Alert {
	//rand.Seed(time.Now().UnixNano())

	rand.NewSource(time.Now().UnixNano())
	idMin := int64(100000000000)
	idMax := int64(999999999999)

	id := rand.Int63n(idMax-idMin) + idMin

	alert := Alert{
		Info: Info{
			ID:           uint64(id),
			Brand:        "Orla",
			Model:        "X1222",
			SerialNumber: "sn-1133n1222",
		},
		Revision: Revision{
			Version:   "1.2",
			CreatedAt: time.Now().AddDate(-1, 0, -12).Unix(),
			UpdateAt:  time.Now().AddDate(0, -4, -5).Unix(),
		},
		Event: Event{
			Severity:   "Critical",
			ReceivedAt: time.Now().Unix(),
			Message:    "Slot 1 critical error",
			Log:        "Some big and complicated log",
		},
	}
	return alert
}
