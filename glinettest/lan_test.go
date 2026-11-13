package glinettest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type fixture struct {
	Request struct {
		ID     int             `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	} `json:"request"`
	Response struct {
		Result any `json:"result"`
		Error  any `json:"error"`
	} `json:"response"`
	name string
}

// TestLANMatchesRouterFixtures replays the static bind fixtures of the real router in the order
// of their request IDs and compares each result and error with the fake router.
func TestLANMatchesRouterFixtures(t *testing.T) {
	paths, err := filepath.Glob("../testdata/fixtures/axt1800_4.7.0/lan.*static_bind*.json")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no fixtures: %v", err)
	}

	var fixtures []fixture
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var f fixture
		if err := json.Unmarshal(data, &f); err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		f.name = filepath.Base(p)
		fixtures = append(fixtures, f)
	}
	slices.SortFunc(fixtures, func(a, b fixture) int { return a.Request.ID - b.Request.ID })

	s := NewServer(t)
	s.sessions["sid"] = struct{}{}

	for _, f := range fixtures {
		params := json.RawMessage(strings.Replace(string(f.Request.Params), `"\u003csid\u003e"`, `"sid"`, 1))
		result, err := s.dispatch(request{Method: f.Request.Method, Params: params})

		var got struct {
			Result any `json:"result"`
			Error  any `json:"error"`
		}
		data, _ := json.Marshal(response{Result: result, Error: asError(err)})
		_ = json.Unmarshal(data, &got)

		if !reflect.DeepEqual(got.Result, f.Response.Result) || !reflect.DeepEqual(got.Error, f.Response.Error) {
			t.Errorf("%s: got result %v error %v, want result %v error %v", f.name, got.Result, got.Error, f.Response.Result, f.Response.Error)
		}
	}
}

func asError(err error) *Error {
	if err == nil {
		return nil
	}
	return err.(*Error)
}
