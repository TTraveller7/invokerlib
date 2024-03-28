package core

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
	functionStates.Created: {
		functionStates.Initialized,
	},
	functionStates.Initialized: {
		functionStates.Running,
		functionStates.Exited,
	},
	functionStates.Running: {
		functionStates.Paused,
		functionStates.Exited,
	},
	functionStates.Paused: {
		functionStates.Running,
		functionStates.Exited,
	},
	functionStates.Exited: {},
}

var (
	// DO NOT directly read this variable. Use getState().
	funcState functionState = functionStates.Created
	stateLock sync.RWMutex

	// DO NOT directory read/write this variable
	transitioning atomic.Bool
)

// getState returns the current function state, and true if is transitioning, false otherwise
func getState() (functionState, bool) {
	stateLock.RLock()
	defer stateLock.RUnlock()
	return funcState, transitioning.Load()
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

	allowedDests := allowedTransitions[funcState]
	for _, allowedDest := range allowedDests {
		if dest == allowedDest {
			funcState = dest
			return nil
		}
	}
	return fmt.Errorf("transit to %s failed: cannot transit from %s to %s", dest, funcState, dest)
}

// startTransition tries to start a transition of current state.
func startTransition(targetState functionState) (func(), error) {
	if swapped := transitioning.CompareAndSwap(false, true); !swapped {
		return nil, fmt.Errorf("another routine is transitioning")
	}

	currState, _ := getState()
	for srcState, destStates := range allowedTransitions {
		if currState != srcState {
			continue
		}
		for _, destState := range destStates {
			if destState == targetState {
				return resetTransition, nil
			}
		}
	}

	resetTransition()
	return nil, fmt.Errorf("transition to target state is not allowed: currState=%v, targetState=%s",
		currState, targetState)
}

func resetTransition() {
	transitioning.Store(false)
}
