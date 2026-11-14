package glinettest

import "encoding/json"

// objectParams decodes params, which must be a JSON object. A null value gives ErrInvalidParams,
// like the router. It is used by the modules that were verified to do this.
func objectParams(params json.RawMessage) (map[string]json.RawMessage, error) {
	var p map[string]json.RawMessage
	if err := json.Unmarshal(params, &p); err != nil || p == nil {
		return nil, ErrInvalidParams
	}
	for _, v := range p {
		if string(v) == "null" {
			return nil, ErrInvalidParams
		}
	}
	return p, nil
}

// truthy returns the value of a flag as the router reads it: no value and false are false,
// and each other value is true, also 0 and "false".
func truthy(raw json.RawMessage) bool {
	return raw != nil && string(raw) != "false"
}
