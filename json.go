package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// AsJSON marshals i to the io.Reader. If i is a function (as determined via
// reflection), it is called and the result is then marshalled to the resulting
// io.Reader - this is useful in the case where the generating the input data is
// expensive to compute or comes from an external source.
func AsJSON(i interface{}) (out io.Reader, err error) {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Func {
		res := v.Call([]reflect.Value{})
		if len(res) != 1 {
			err = fmt.Errorf("callable arguments to AsJSON must return a single value")
			return
		}
		i = res[0].Interface()
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	err = enc.Encode(i)
	if err != nil {
		err = fmt.Errorf("failed to encode snapshot as JSON: %w", err)
	}
	out = buf
	return
}

// WithCreateSnapshotAsJSON configures GetTestInput to use AsJSON as the
// CreateSnapshot and sets the file extension to ".json".
func WithCreateSnapshotAsJSON(i interface{}) GetTestInputOption {
	return func(o *GetTestInputOptions) {
		o.CreateSnapshot = func() (io.Reader, error) { return AsJSON(i) }
		o.FileExtension = ".json"
	}
}
