package glinet

import "fmt"

// flag is a bool that the router sends and accepts as the number 0 or 1. It also decodes a JSON
// bool. Another number gives an error, so that a value with a third state is not read as a bool.
type flag bool

func (f flag) MarshalJSON() ([]byte, error) {
	if f {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

func (f *flag) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "null":
		return nil
	case "1", "true":
		*f = true
		return nil
	case "0", "false":
		*f = false
		return nil
	}
	return fmt.Errorf("glinet: value is not 0, 1 or a bool: %s", data)
}
