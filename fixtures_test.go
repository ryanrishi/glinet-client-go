package glinet_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

// fixtureResult returns the result in the fixture name of the test router.
func fixtureResult(t *testing.T, name string) json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata/fixtures/axt1800_4.7.0", name))
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Response struct {
			Result json.RawMessage `json:"result"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return f.Response.Result
}

// checkLossless checks that the JSON encoding of got has each value of result, a JSON router
// result, at the same path. Object keys in omitted are not checked.
func checkLossless(t *testing.T, got any, result json.RawMessage, omitted ...string) {
	t.Helper()
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var gotValue, want any
	if err := json.Unmarshal(data, &gotValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(result, &want); err != nil {
		t.Fatal(err)
	}
	for _, diff := range lossDiffs("", gotValue, want, omitted) {
		t.Error(diff)
	}
}

func lossDiffs(path string, got, want any, omitted []string) []string {
	switch want := want.(type) {
	case map[string]any:
		gotMap, ok := got.(map[string]any)
		if !ok {
			return []string{fmt.Sprintf("%s: got %v, want an object", path, got)}
		}
		var diffs []string
		for k, v := range want {
			if slices.Contains(omitted, k) {
				continue
			}
			g, ok := gotMap[k]
			if !ok {
				diffs = append(diffs, fmt.Sprintf("%s.%s: missing", path, k))
				continue
			}
			diffs = append(diffs, lossDiffs(path+"."+k, g, v, omitted)...)
		}
		return diffs
	case []any:
		gotSlice, ok := got.([]any)
		if !ok || len(gotSlice) != len(want) {
			return []string{fmt.Sprintf("%s: got %v, want %d items", path, got, len(want))}
		}
		var diffs []string
		for i := range want {
			diffs = append(diffs, lossDiffs(fmt.Sprintf("%s[%d]", path, i), gotSlice[i], want[i], omitted)...)
		}
		return diffs
	default:
		if !reflect.DeepEqual(got, want) {
			return []string{fmt.Sprintf("%s: got %v, want %v", path, got, want)}
		}
		return nil
	}
}
