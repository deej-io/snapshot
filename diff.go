package snapshot

import (
	"fmt"
	"io"

	"github.com/google/go-cmp/cmp"
)

// StringDiffComparator reads expected and actual into strings compares them using
// cmp.Diff from the go-cmp library. On failure the diff string is returned.
func StringDiffComparator(expected, actual io.Reader) (ok bool, msg string) {
	actualString, err := readToString(actual)
	if err != nil {
		msg = fmt.Sprintf("failed to read actual io.Reader: %v", err.Error())
		return
	}
	expectedString, err := readToString(expected)
	if err != nil {
		msg = fmt.Sprintf("failed to read expected io.Reader: %v", err.Error())
		return
	}
	msg = cmp.Diff(expectedString, actualString)
	ok = msg == ""
	return
}
