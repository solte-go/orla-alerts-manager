package model

import "encoding/json"

type Alert struct {
	// ID       string   `json:"hash_id" bson:"_id"`
	Info     Info     `json,bson:"info"`
	Revision Revision `json,bson:"revision"`
	Event    Event    `json,bson:"event"`
}

type Info struct {
	ID           uint64 `json,bson:"id"`
	Name         string `json,bson:"name"`
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
	Severity       string `json,bson:"severity"`
	ReceivedAt     int64  `json,bson:"time_of_event"`
	AcknowledgedAt int64  `json,bson:"acknowledged"`
	ClosedAt       int64  `json,bson:"closed_at"`
	Message        string `json,bson:"message"`
	Log            string `json,bson:"log"`
}

func (a *Alert) ToJSON() []byte {
	res, _ := json.Marshal(a)
	return res
}
