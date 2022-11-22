package glinet

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

// SinglePayloadis used to represent a generic JSONRPC payload where one result (Node) is included as an {} in the "result" key
type SinglePayload struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Result  *Node  `json:"result"`
}

// ManyPayload is used to represent a generic JSONRPC payload where many results (Nodes) are included in an [] in the "result" key
type ManyPayload struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      int64   `json:"id"`
	Result  *[]Node `json:"result"`
}

// Node is used to represent a generic JSONRPC resource
type Node struct {
}

// largely copy pasta from https://github.com/google/jsonapi/blob/master/request.go
func UnmarshalPayload(in io.Reader, model interface{}) error {
	payload := new(SinglePayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return err
	}

	return nil
}

func UnmarshalManyPayload(in io.Reader, t reflect.Type) ([]interface{}, error) {
	payload := new(ManyPayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, err
	}

	return nil, errors.New("unimplemented")
}

func unmarshalNode(data *Node, model reflect.Value) (err error) {
	modelValue := model.Elem()
	modelType := model.Type()

	for i := 0; i < modelValue.NumField(); i++ {
		fieldType := modelType.Field(i)
		tag := fieldType.Tag.Get("jsonrpc")
		if tag == "" {
			continue
		}

		//fieldValue := modelValue.Field(i)
	}

	return errors.New("unimplemented")
}
