package model

import (
	"time"
)

type TimeStamp struct {
	int64
}

func (ts TimeStamp) MarshalJSON() ([]byte, error) {
	t := time.Unix(ts.int64, 0)
	str := t.Format("2006-01-02T15:04:05-07:00")
	return []byte("\"" + str + "\""), nil
}

func TimeStampNow() TimeStamp {
	return TimeStamp{time.Now().UTC().Unix()}
}
