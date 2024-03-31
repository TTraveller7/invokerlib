package consts

import "fmt"

var (
	ErrProcessorTypeNotRecognized = fmt.Errorf("unrecognized processor type")
	ErrInputProcessorNotComplete  = fmt.Errorf("left input processor and right input processor must not be empty")
	ErrConfigNotInitialized       = fmt.Errorf("config is not initialized")
	ErrReadPtrSameAsWritePtr      = fmt.Errorf("read ptr is same as write ptr")
	ErrLockBufferFailure          = fmt.Errorf("fail to lock buffer")
	ErrLockSlotFailure            = fmt.Errorf("fail to lock slot")
)

func ErrKakfaAddressEmpty(prefix string) error {
	return fmt.Errorf("%s: kafka address cannot be empty", prefix)
}

func ErrKafkaTopicEmpty(prefix string) error {
	return fmt.Errorf("%s: kafka address cannot be empty", prefix)
}
