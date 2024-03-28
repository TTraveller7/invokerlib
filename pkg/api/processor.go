package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/TTraveller7/invokerlib/pkg/conf"
	"github.com/TTraveller7/invokerlib/pkg/consts"
	"github.com/TTraveller7/invokerlib/pkg/core"
	"github.com/TTraveller7/invokerlib/pkg/logs"
	"github.com/TTraveller7/invokerlib/pkg/models"
	"github.com/TTraveller7/invokerlib/pkg/state"
	"github.com/TTraveller7/invokerlib/pkg/utils"
)

var ProcessorCommands = struct {
	Initialize string
	Run        string
	Exit       string
	Ping       string
	Cat        string
}{
	Initialize: "initialize",
	Run:        "run",
	Exit:       "exit",
	Ping:       "ping",
	Cat:        "cat",
}

func ProcessorHandle(w http.ResponseWriter, r *http.Request, pc *models.ProcessorCallbacks) {
	switch r.URL.Path {
	case "/metrics":
		utils.MetricsHandler().ServeHTTP(w, r)
		return
	}

	resp := &InvokerResponse{}
	var err error
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%v. %s", panicErr, string(debug.Stack()))
		}
		if err != nil {
			resp = failureResponse(err)
		}
		respBytes, _ := json.Marshal(resp)
		w.Write(respBytes)
	}()

	content, readAllErr := io.ReadAll(r.Body)
	if readAllErr != nil {
		err = fmt.Errorf("read request body failed: %v", readAllErr)
		logs.Printf("%v", err)
		return
	}

	req := &InvokerRequest{}
	if unmarshalErr := json.Unmarshal(content, req); unmarshalErr != nil {
		err = fmt.Errorf("unmarshal request failed: %v", unmarshalErr)
		logs.Printf("%v", err)
		return
	}
	if req.Command == "" {
		err = fmt.Errorf("request command is missing")
		logs.Printf("%v", err)
		return
	}
	logs.Printf("received request: %s", string(utils.SafeJsonIndent(req)))

	var handleErr error
	switch req.Command {
	case ProcessorCommands.Initialize:
		resp, handleErr = handleInitialize(req, pc)
	case ProcessorCommands.Ping:
		resp = successResponse()
		resp.Message = "pong"
	case ProcessorCommands.Run:
		resp, handleErr = handleRun()
	case ProcessorCommands.Cat:
		resp, handleErr = handleCat()
	default:
		err = fmt.Errorf("unrecognized command %s", req.Command)
		logs.Printf("%v", err)
		return
	}
	if handleErr != nil {
		err = fmt.Errorf("handle processor command failed: %v", handleErr)
		return
	}
}

func handleInitialize(req *InvokerRequest, pc *models.ProcessorCallbacks) (*InvokerResponse, error) {
	logs.Printf("handle initialize starts")
	ipc := &conf.InternalProcessorConfig{}
	if err := UnmarshalParams(req.Params, ipc); err != nil {
		err = fmt.Errorf("unmarshal params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	if err := core.Initialize(ipc, pc); err != nil {
		err = fmt.Errorf("Initialize failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	logs.Printf("handle initialize finished")
	return successResponse(), nil
}

func handleRun() (*InvokerResponse, error) {
	logs.Printf("handle run starts")
	if err := core.Run(); err != nil {
		err = fmt.Errorf("Run failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	logs.Printf("handle run finished")
	return successResponse(), nil
}

func handleCat() (*InvokerResponse, error) {
	logs.Printf("handle cat starts")
	res, err := cat(context.Background())
	if err != nil {
		err = fmt.Errorf("cat failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	msgBytes, err := json.Marshal(res)
	if err != nil {
		err = fmt.Errorf("marshal cat result failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	resp := successResponse()
	resp.Message = string(msgBytes)
	logs.Printf("handle cat finished")
	return resp, nil
}

func cat(ctx context.Context) (ProcessorCatResult, error) {
	limit := consts.DefaultCatLimit

	resp := make(map[string][]StateStoreEntry, 0)
	for name, stateStore := range stateStores {
		keys, err := stateStore.Keys(ctx, limit)
		if err != nil {
			return nil, err
		}
		entries := make([]StateStoreEntry, 0)
		for _, key := range keys {
			val, err := stateStore.Get(ctx, key)
			if err != nil {
				return nil, err
			}
			entries = append(entries, StateStoreEntry{
				Key: key,
				Val: string(val),
			})
		}
		resp[name] = entries
	}
	return resp, nil
}

var stateStores map[string]state.StateStore = make(map[string]state.StateStore, 0)

func AddStateStore(name string, stateStore state.StateStore) {
	stateStores[name] = stateStore
}
