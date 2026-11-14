package glinettest

import (
	"encoding/json"
	"errors"
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
		Result     any `json:"result"`
		Error      any `json:"error"`
		HTTPStatus int `json:"http_status"`
	} `json:"response"`
	name string
}

// replayFixtures replays the fixtures of the real router that match pattern in the order of their
// request IDs and compares each result, error and HTTP status 500 with the fake router.
func replayFixtures(t *testing.T, pattern string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("../testdata/fixtures/axt1800_4.7.0", pattern))
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

		if f.Response.HTTPStatus != 0 {
			if f.Response.HTTPStatus != 500 || !errors.Is(err, errInternalServerError) {
				t.Errorf("%s: got result %v error %v, want HTTP status %d", f.name, result, err, f.Response.HTTPStatus)
			}
			continue
		}

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
