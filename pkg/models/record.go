package models

import (
	"time"

	"github.com/IBM/sarama"
)

type Record struct {
	msgTimestamp time.Time
	key          string
	value        []byte
}

func NewRecord(key string, value []byte) *Record {
	return &Record{
		key:   key,
		value: value,
	}
}

func NewRecordWithConsumerMessage(msg *sarama.ConsumerMessage) *Record {
	return &Record{
		msgTimestamp: msg.Timestamp,
		key:          string(msg.Key),
		value:        msg.Value,
	}
}

func (r *Record) Key() string {
	return r.key
}

func (r *Record) Value() []byte {
	return r.value
}
