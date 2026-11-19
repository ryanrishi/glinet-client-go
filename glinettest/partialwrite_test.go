package glinettest

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

// checkConcurrentPartialWrites runs one goroutine for each field. Each goroutine writes only its
// field, with each of its values in turn, and after each write reads the state. The state must
// then have the value of that write: a partial write of another field must not set an old value.
func checkConcurrentPartialWrites(t *testing.T, write Method, read func() any, values map[string][]string) {
	t.Helper()
	var wg sync.WaitGroup
	for field, fieldValues := range values {
		wg.Go(func() {
			for _, v := range fieldValues {
				if _, err := write(json.RawMessage(fmt.Sprintf(`{%q:%s}`, field, v))); err != nil {
					t.Errorf("write %s %s: %v", field, v, err)
					return
				}
				var state map[string]json.RawMessage
				mustUnmarshal(string(mustMarshal(read())), &state)
				if got := string(state[field]); got != v {
					t.Errorf("after write %s %s: %s = %s (lost update)", field, v, field, got)
					return
				}
			}
		})
	}
	wg.Wait()
}

func numbers(n int) []string {
	res := make([]string, n)
	for i := range res {
		res[i] = fmt.Sprint(i + 1)
	}
	return res
}

func toggles(n int) []string {
	res := make([]string, n)
	for i := range res {
		res[i] = fmt.Sprint(i%2 == 0)
	}
	return res
}
