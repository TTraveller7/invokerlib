package core

import "fmt"

var (
	ErrConfigNotInitialized    = fmt.Errorf("config is not initialized")
	ErrRedisConfNotInitialized = fmt.Errorf("redis config is not initialized")
	ErrConsumerNotify          = fmt.Errorf("consumer exits with notify")
)
