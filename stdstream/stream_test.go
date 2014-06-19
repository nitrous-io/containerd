package stdstream

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func contain(got []data, item data) bool {
	for _, g := range got {
		if reflect.DeepEqual(g, item) {
			return true
		}
	}
	return false
}

// assert arriving data that can be out of order
func assertData(t *testing.T, output string, expected []data) {
	msgs := strings.Split(output, "\n")
	if len(msgs) != len(expected) {
		t.Fatalf("Expected %v to have %v items", msgs, len(expected))
	}
	got := make([]data, len(msgs))
	for i, msg := range msgs {
		json.Unmarshal([]byte(msg), &got[i])
	}
	for _, e := range expected {
		if !contain(got, e) {
			t.Fatalf("Expected item %v to be in %v", e, got)
		}
	}
}

func TestStream(t *testing.T) {
	output := bytes.NewBuffer(nil)
	stdout := strings.NewReader("helloworld")
	stderr := strings.NewReader("newerror")

	w := NewStdStream(output, map[string]io.Reader{
		"stdout": stdout,
		"stderr": stderr,
	})

	errDone := make(chan error, 1)
	go func() {
		errDone <- w.Flush()
	}()

	after := time.After(10 * time.Millisecond)
	select {
	case <-after:
		t.Fatal("timed out")
	case err := <-errDone:
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		strout := strings.TrimSpace(output.String())
		assertData(t, strout, []data{
			{Type: "stdout", Data: []byte("helloworld")},
			{Type: "stderr", Data: []byte("newerror")},
		})
	}
}
