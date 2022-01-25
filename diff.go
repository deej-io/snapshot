package snapshot

import (
	"fmt"
	"io"

	"github.com/google/go-cmp/cmp"
)

// CmpDiffComparator reads expected and actual into strings compares them using
// cmp.Diff from the go-cmp library. On failure the diff string is returned.
func CmpDiffComparator(expected, actual io.Reader) (ok bool, msg string) {
	actualBytes, err := readToString(actual)
	if err != nil {
		msg = fmt.Sprintf("failed to read actual io.Reader: %v", err.Error())
		return
	}
	expectedBytes, err := io.ReadAll(expected)
	if err != nil {
		msg = fmt.Sprintf("failed to read expected io.Reader: %v", err.Error())
		return
	}
	msg = cmp.Diff(expectedBytes, actualBytes)
	ok = msg == ""
	return
}
