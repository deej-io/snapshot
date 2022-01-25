package snapshot

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func getInputOutputPathsAndClean(t *testing.T) (input, output string) {
	input = getSnapshotFilePath(t, "input", ".txt")
	output = getSnapshotFilePath(t, "output", ".txt")
	if err := os.RemoveAll(filepath.Dir(input)); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove test snapshots: %v", err)
	}
	return
}

func TestRoundTrip(t *testing.T) {
	inputP, outputP := getInputOutputPathsAndClean(t)
	input := GetTestInput(t, WithInputSnapshotReader(strings.NewReader("hello")))
	ok, msg := Match(t, input)

	if !ok {
		t.Fatalf("expectd match but didn't: %v", msg)
	}
	inputF, err := os.Open(inputP)
	if err != nil {
		t.Fatalf("failed to open input snapshot: %v", err)
	}
	if str := readToStringUnchecked(inputF); str != "hello" {
		t.Fatalf("unexpected input snapshot. expected %q, got %q", "hello", str)
	}
	outputF, err := os.Open(outputP)
	if err != nil {
		t.Fatalf("failed to open output snapshot: %v", err)
	}
	if str := readToStringUnchecked(outputF); str != "hello" {
		t.Fatalf("unexpected output snapshot. expected %q, got %q", "hello", str)
	}
}

func TestOutputChanged(t *testing.T) {
	_, _ = getInputOutputPathsAndClean(t)
	firstOutput := strings.NewReader("hello")
	if ok, msg := Match(t, firstOutput); !ok {
		t.Errorf("expected first match to succeed: %v", msg)
	}
	firstOutput = strings.NewReader("hello")
	if ok, msg := Match(t, firstOutput); !ok {
		t.Errorf("expected first match to succeed a second time: %v", msg)
	}
	secondOutput := strings.NewReader("world")
	expectedMsg := `expected "hello", got "world"`
	if ok, msg := Match(t, secondOutput); ok || msg != expectedMsg {
		t.Errorf("expected second match to be false, got: %v", ok)
		t.Errorf("expected message to be %q, got %q", expectedMsg, msg)
	}
}

func readToStringUnchecked(r io.Reader) (s string) {
	s, _ = readToString(r)
	return
}

func TestInputChanged(t *testing.T) {
	_, _ = getInputOutputPathsAndClean(t)
	firstInput := GetTestInput(t, WithInputSnapshotReader(strings.NewReader("hello")))
	if str := readToStringUnchecked(firstInput); str != "hello" {
		t.Errorf("expected %q, got %q", "hello", str)
	}
	secondInput := GetTestInput(t, WithInputSnapshotReader(strings.NewReader("world")))
	if str := readToStringUnchecked(secondInput); str != "hello" {
		t.Errorf("expected %q, got %q", "hello", str)
	}
}
