package models

import (
	"github.com/IBM/sarama"
)

type Record struct {
	msg *sarama.ConsumerMessage
}

func NewRecord(msg *sarama.ConsumerMessage) *Record {
	return &Record{
		msg: msg,
	}
}

func (r *Record) Key() []byte {
	return r.msg.Key
}

func (r *Record) Value() []byte {
	return r.msg.Value
}
