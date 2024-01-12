package invokerlib

import "encoding/json"

type InvokerRequest struct {
	Command string         `json:"command"`
	Params  map[string]any `json:"params"`
}

func UnmarshalParams(params map[string]any, dest any) error {
	marshalledParams, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshalledParams, dest)
}

func MarshalToParams(s any) (map[string]any, error) {
	sBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any, 0)
	if json.Unmarshal(sBytes, &params); err != nil {
		return nil, err
	}
	return params, nil
}
