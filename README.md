# snapshot - Snapshot test utilities for go

snapshot provides a set of utility functions for creating and loading snapshot files for using snapshot tests.

## Install

```
go get -u github.com/djrollins/snapshot
```

## Output Snapshot Example

```go
package example

import (
	"testing"

	"github.com/djrollins/snapshot"
)

func ScanSum(in []int) (out []int) {
	sum := 0
	for i := range in {
		sum += i
		out = append(out, sum)
	}
	return
}

func TestScanPlus(t *testing.T) {
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i
	}
	scan := ScanSum(nums)
	asJSON, err := snapshot.AsJSON(scan)
	if err != nil {
		t.Fatalf("failed to encode scan result to JSON")
	}
	ok, msg := snapshot.Match(t, asJSON)
	if !ok {
		t.Fatalf("scan did not match snapshot: %v", msg)
	}
}
```

## Input Snapshot Example

```go
package example

import (
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/djrollins/snapshot"
)

var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EAR == "ExpensiveAsyncronousRandom" opertion provided by an external system,
// which also exposes a version number for their library.
const EARVersion = "v0.2.0"

func ExpensiveAsyncronousRandomOperation() (out io.Reader, err error) {
	out, in := io.Pipe()
	go func() {
		for i := 1; i < 1000; i++ {
			n := rand.Intn(len(letters))
			c := letters[n]
			in.Write([]byte{c})
			time.Sleep(2 * time.Millisecond)
		}
		in.Close()
	}()
	return
}

func TestEAR(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	// First time this test runs, this will take ~2 seconds. However, the
	// the input data is persisted to disk and reused on subsequent runs. This
	// allows for fast, repeatable tests that require complex input data
	// without having to manually create it. This should be used with custom
	// snapshot names based on versions so that a new snapshot is created
	// and breaking changes can be detected between versions.
	input := snapshot.GetTestInput(t,
		snapshot.WithCreateSnapshot(ExpensiveAsyncronousRandomOperation),
		snapshot.WithInputSnapshotName("ear_"+EARVersion),
	)
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, input)

	upper := strings.ToUpper(buf.String())

	// Match will pass on the first run, as there is no snapshot recorded,
	// however it will ensure that future runs result in the same data,
	// until the snapshot is deleted.
	ok, msg := snapshot.Match(t, strings.NewReader(upper))
	if !ok {
		t.Fatalf("unexpected output: %v", msg)
	}
}
```
