package jobs

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
)

var (
	NoJobResponseBytes      []byte
	UnknownRequestTypeBytes []byte
	InvalidMessageTypeBytes []byte
	EmptyOkResponseBytes    []byte
)

func init() {
	NoJobResponseBytes, _ = serializeMessage(&ErrorResponse{Status: "no-job"})
	UnknownRequestTypeBytes, _ = serializeMessage(&ErrorResponse{Status: "error", Error: "unknown request type"})
	InvalidMessageTypeBytes, _ = serializeMessage(&ErrorResponse{Status: "error", Error: "unable to decode message"})
	EmptyOkResponseBytes, _ = serializeMessage(&ErrorResponse{Status: "ok"})
}

type Base struct {
	Request string `json:"request"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type PutRequest struct {
	Request  string `json:"request"`
	Queue    string `json:"queue"`
	Priority int    `json:"pri"`
	Job      any    `json:"job"`
}

type PutResponse struct {
	Status string `json:"status"`
	Id     uint64 `json:"id"`
}

type GetRequest struct {
	Request string   `json:"request"`
	Queues  []string `json:"queues"`
	Wait    bool     `json:"wait"`
}

type GetResponse struct {
	Status   string `json:"status"`
	Id       uint64 `json:"id"`
	Job      any    `json:"job"`
	Priority int    `json:"pri"`
	Queue    string `json:"queue"`
}

type DeleteRequest struct {
	Request string `json:"request"`
	Id      uint64 `json:"id"`
}

type AbortRequest struct {
	Request string `json:"request"`
	Id      uint64 `json:"id"`
}

func serializeMessage(msg any) ([]byte, error) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, '\n')
	return bytes, nil
}

func decodeMessage(x []byte) (any, error) {
	var base Base
	err := json.Unmarshal(x, &base)
	if err != nil {
		return nil, err
	}

	// we do this twice since i don't want to do the effort of doing it better
	switch base.Request {
	case "put":
		var p PutRequest
		err = json.Unmarshal(x, &p)
		return &p, err
	case "get":
		var p GetRequest
		err = json.Unmarshal(x, &p)
		return &p, err
	case "delete":
		var p DeleteRequest
		err = json.Unmarshal(x, &p)
		return &p, err
	case "abort":
		var p AbortRequest
		err = json.Unmarshal(x, &p)
		return &p, err
	default:
		log.Warn().Str("request_type", base.Request).Msg("unknown request type")
		return nil, errors.New("unknown request type")
	}
}
