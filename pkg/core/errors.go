package core

import "fmt"

var (
	ErrRedisConfNotInitialized = fmt.Errorf("redis config is not initialized")
	ErrConsumerNotify          = fmt.Errorf("consumer exits with notify")
)
