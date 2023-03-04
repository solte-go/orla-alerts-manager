package model

import (
	"math/rand"
	"time"
)

func GetTestAlert() Alert {

	idMin := int64(800)
	idMax := int64(2147483645)

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
			CreatedAt: int64(time.Now().AddDate(-1, 0, -12).Unix()),
			UpdateAt:  int64(time.Now().AddDate(0, -4, -5).Unix()),
		},
		Event: Event{
			Severity:    "Critical",
			TimeOfEvent: time.Now().Unix(),
			Message:     "Slot 1 critical error",
			Log:         "Some big and complicated log",
		},
	}

	alert.CreateHashID()
	return alert
}