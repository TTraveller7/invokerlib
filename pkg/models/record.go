package models

import (
	"time"

	"github.com/IBM/sarama"
)

type Record struct {
	msgTimestamp time.Time
	key          []byte
	value        []byte
}

func NewRecord(key, value []byte) *Record {
	return &Record{
		key:   key,
		value: value,
	}
}

func NewRecordWithConsumerMessage(msg *sarama.ConsumerMessage) *Record {
	return &Record{
		msgTimestamp: msg.Timestamp,
		key:          msg.Key,
		value:        msg.Value,
	}
}

func (r *Record) Key() []byte {
	return r.key
}

func (r *Record) Value() []byte {
	return r.value
}
