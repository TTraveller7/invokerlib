package invokerlib

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type functionState string

var functionStates = struct {
	Created     functionState
	Initialized functionState
	Running     functionState
	Paused      functionState
	Exited      functionState
}{
	Created:     "Created",
	Initialized: "Initialized",
	Running:     "Running",
	Paused:      "Paused",
	Exited:      "Exited",
}

var allowedTransitions = map[functionState][]functionState{
	functionStates.Created: []functionState{
		functionStates.Initialized,
	},
	functionStates.Initialized: []functionState{
		functionStates.Running,
	},
	functionStates.Running: []functionState{
		functionStates.Paused,
		functionStates.Exited,
	},
	functionStates.Paused: []functionState{
		functionStates.Running,
		functionStates.Exited,
	},
	functionStates.Exited: []functionState{},
}

var (
	// DO NOT directly read this variable
	state     functionState = functionStates.Created
	stateLock sync.RWMutex

	// DO NOT directory read/write this variable
	transitioning atomic.Bool
)

// getState returns the current function state, and true if is transitioning, false otherwise
func getState() (functionState, bool) {
	stateLock.RLock()
	defer stateLock.RUnlock()
	return state, transitioning.Load()
}

func transitToInitialized() error {
	return transitTo(functionStates.Initialized)
}

func transitToRunning() error {
	return transitTo(functionStates.Running)
}

func transitToPaused() error {
	return transitTo(functionStates.Paused)
}

func transitToExited() error {
	return transitTo(functionStates.Exited)
}

func transitTo(dest functionState) error {
	if !transitioning.Load() {
		return fmt.Errorf("another routine is transitioning")
	}

	stateLock.Lock()
	defer stateLock.Unlock()

	allowedDests := allowedTransitions[state]
	for _, allowedDest := range allowedDests {
		if dest == allowedDest {
			state = dest
			return nil
		}
	}
	return fmt.Errorf("transit to %s failed: cannot transit from %s to %s", dest, state, dest)
}

// startTransition tries to start a transition of current state.
func startTransition() (func(), error) {
	if swapped := transitioning.CompareAndSwap(false, true); !swapped {
		return nil, fmt.Errorf("another routine is transitioning")
	} else {
		return resetTransition, nil
	}
}

func resetTransition() {
	transitioning.Store(false)
}

func isInitialized() bool {
	s, transitioning := getState()
	if transitioning {
		return false
	} else {
		return s == functionStates.Initialized
	}
}
