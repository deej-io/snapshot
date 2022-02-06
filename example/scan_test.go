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
