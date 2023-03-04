package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Alert struct {
	HashID   string   `json:"hash_id" bson:"_id"`
	Info     Info     `json,bson:"info"`
	Revision Revision `json,bson:"revision"`
	Event    Event    `json,bson:"event"`
}

func (a *Alert) CreateHashID() {
	hash := sha256.New()
	_, _ = fmt.Fprint(hash, a.Info.ID, a.Revision.Version, a.Event.TimeOfEvent)
	a.HashID = hex.EncodeToString(hash.Sum(nil))
}

type Info struct {
	ID           uint64 `json,bson:"id"`
	Brand        string `json,bson:"brand"`
	Type         string `json,bson:"type"`
	Model        string `json,bson:"model"`
	SerialNumber string `json,bson:"serial_number"`
}

type Revision struct {
	Version        string `json,bson:"version"`
	CreatedAt      int64  `json,bson:"created_at"`
	UpdateAt       int64  `json,bson:"update_at"`
	DecommissionAt int64  `json,bson:"decommission_at"`
}

type Event struct {
	Severity    string `json,bson:"severity"`
	TimeOfEvent int64  `json,bson:"time_of_event"`
	Message     string `json,bson:"message"`
	Log         string `json,bson:"log"`
}
