// Package snapshot provides utility functions for creating and loading on-disk
// data for use in snapshot tests.
package snapshot

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// getSnapshotFilePath returns a path to file named name.ext  located within
// directory __snapshots__ directory at the same level as the file that
// contains the currently running test. This directory is named after the test
// name and is therefore unique to each test.
func getSnapshotFilePath(t *testing.T, name, ext string) string {
	_, file, _, _ := runtime.Caller(2)
	return filepath.Join(filepath.Dir(file), "__snapshots__", t.Name(), name+ext)
}

// A SnapshotCreator is a function that can be provided to GetTestInput which
// will be used used in the case where an input snapshot file does not exist.
type SnapshotCreator func() (io.Reader, error)

// GetTestInputOptions are the set of options to configure the behaviour of
// GetTestInput.
type GetTestInputOptions struct {
	// SnapshotName is the name of the snapshot to load, without the
	// extension. This defaults to the value "input".
	SnapshotName string
	// FileExtension is the file extension of the snapshot. This defaults
	// to ".txt".
	FileExtension string
	// CreateSnapshot will be called if the snapshot file does not exist
	// The resulting io.Reader provides the input data for the currently
	// running test, as well as being copied to disk for future tests.
	// This is useful in cases where input data may be volatile or random
	// and would therefore usually be unsuitable for snapshot tests.
	CreateSnapshot SnapshotCreator
}

// GetTestInputOption may be an argument to GetTestInput in order to change
// GetTestInputOptions
type GetTestInputOption func(*GetTestInputOptions)

// WithInputSnapshotName overrides the snapshot name. This is useful in cases
// where there may be multiple input snapshots for a test. This option does not
// change the file extension.
func WithInputSnapshotName(name string) GetTestInputOption {
	return func(o *GetTestInputOptions) {
		o.SnapshotName = name
	}
}

// WithInputSnapshotFileExtension overrides the file extension for the
// resulting snapshot file. This is useful if the snapshots are read by
// external tools and make use of the extensions to determine the filetype.
func WithInputSnapshotFileExtension(ext string) GetTestInputOption {
	return func(o *GetTestInputOptions) {
		o.FileExtension = ext
	}
}

// WithSnapshotFilename overrides snapshot name and file extension of the
// resulting snapshot file.
func WithSnapshotFilename(filename string) GetTestInputOption {
	return func(o *GetTestInputOptions) {
		o.FileExtension = filepath.Ext(filename)
		o.SnapshotName = strings.TrimSuffix(filename, o.FileExtension)
	}
}

// WithCreateSnapshot provides a SnapshotCreator function to specify the
// data for the test when no snapshot file exists. This data is also persisted
// to disk and used for subsequent test runs.
func WithCreateSnapshot(r func() (io.Reader, error)) GetTestInputOption {
	return func(o *GetTestInputOptions) {
		o.CreateSnapshot = r
	}
}

// WithInputSnapshotReader creates SnapshotCreator func from a reader.
func WithInputSnapshotReader(r io.Reader) GetTestInputOption {
	return WithCreateSnapshot(func() (io.Reader, error) { return r, nil })
}

// GetTestInput gets loads the input snapshot for a particular test case.  By
// default, this looks for the file at the location
// <test-directory>/__snapshots__/<test-name>/input.txt. The base directory
// cannot be changed, however the file name can be overridden using with the
// relevant GetTestInputOption arguments.  If the input snapshot file does not
// exist, the test immediately fails, unless a SnapshotCreator is provided
// as an argument. In this case the resulting reader for the SnapshotCreator is
// used as the input data for the current test run and persisted to disk for
// use in subsequent test runs.
func GetTestInput(t *testing.T, optFns ...GetTestInputOption) (out io.Reader) {
	opts := GetTestInputOptions{
		SnapshotName:   "input",
		FileExtension:  ".txt",
		CreateSnapshot: nil,
	}
	for _, optFn := range optFns {
		optFn(&opts)
	}

	p := filepath.Clean(getSnapshotFilePath(t, opts.SnapshotName, opts.FileExtension))
	file, err := os.Open(p)
	t.Logf("input snapshot filename: %v", p)
	if err == nil {
		t.Cleanup(func() { _ = file.Close() })
		t.Logf("using existing snapshot")
		out = io.NopCloser(file)
		return
	}
	if os.IsNotExist(err) {
		if opts.CreateSnapshot == nil {
			t.Fatalf("snapshot file %q does not exist and no CreateSnapshot option was provided", p)
		}
		in, err := opts.CreateSnapshot()
		if err != nil {
			t.Fatalf("snapshot creator failed with an error %v", err)
		}
		t.Log("creating new input snapshot")
		err = os.MkdirAll(filepath.Dir(p), 0750)
		if err != nil {
			t.Fatalf("failed to create input snapshot file %v: %v", p, err.Error())
		}
		file, err := os.Create(p)
		if err != nil {
			t.Fatalf("failed to open newly created snapshot file: %v: %v", p, err.Error())
		}
		t.Cleanup(func() { _ = file.Close() })
		out = io.TeeReader(in, file)
	} else {
		t.Fatalf("error opening input snapshot file")
	}
	return
}

// A Comparator can be used to override the default snapshot comparison. This
// function should compare the expected and actual io.Readers and return ok as
// true if they are deemed equal. The return value msg should be a human
// readable reason for the failure in the case where the values are not equal.
type Comparator func(expected, actual io.Reader) (ok bool, msg string)

// A ReaderNormaliser is a function that takes an io.Reader and returns a new
// io.Reader after some processing. This function is applied to the actual and
// expected io.Readers before they are passed to the Comparator
type ReaderNormaliser func(io.Reader) io.Reader

// MatchOptions are the set of options to configure the behaviour of Match.
type MatchOptions struct {
	// SnapshotName is the name of the snapshot to load, without the
	// extension. This defaults to the value "output".
	SnapshotName string
	// FileExtension is the file extension of the snapshot. This defaults
	// to ".txt".
	FileExtension string
	// Comparator is a function to compare the actual and expected
	// io.Readers.
	Comparator Comparator
	// ReaderNormaliser is applied to the actual and expected io.Readers before
	// being passed to the comparator. This can be used to perform some clean
	// or modifications (i.e. sorting) of the snapshot/actual data before
	// comparison.
	ReaderNormaliser ReaderNormaliser
}

// MatchOption may be an argument to Match in order to change MatchOptions.
type MatchOption func(*MatchOptions)

// WithOutputSnapshotName overrides the snapshot name. This is useful in cases
// where there may be multiple output snapshots for a test. This option does not
// change the file extension.
func WithOutputSnapshotName(name string) MatchOption {
	return func(o *MatchOptions) {
		o.SnapshotName = name
	}
}

// WithOutputSnapshotFileExtension overrides the file extension for the
// resulting snapshot file. This is useful if the snapshots are read by
// external tools and make use of the extensions to determine the filetype.
func WithOutputSnapshotFileExtension(ext string) MatchOption {
	return func(o *MatchOptions) {
		o.FileExtension = ext
	}
}

// WithComparator overrides the default comparator with a custom function.
func WithComparator(cmp Comparator) MatchOption {
	return func(o *MatchOptions) {
		o.Comparator = cmp
	}
}

// WithReaderNormaliser provides a ReaderNormaliser function to apply to the
// actual and expected io.Readers before comparison.
func WithReaderNormaliser(rn ReaderNormaliser) MatchOption {
	return func(o *MatchOptions) {
		o.ReaderNormaliser = rn
	}
}

// readToString reads r into a string.
func readToString(r io.Reader) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, r)
	return buf.String(), err
}

// StringComparator reads expected and actual into strings and performs an
// equality check.
func StringComparator(expected, actual io.Reader) (ok bool, msg string) {
	eStr, err := readToString(expected)
	if err != nil {
		msg = "failed to read expected data from reader: " + err.Error()
		return
	}
	aStr, err := readToString(actual)
	if err != nil {
		msg = "failed to read actual data from reader: " + err.Error()
		return
	}
	ok = eStr == aStr
	if !ok {
		msg = fmt.Sprintf("expected %q, got %q", eStr, aStr)
	}
	return
}

// NopReaderNormaliser is the default ReaderNormaliser. It passes the input
// io.Reader through unmodified
func NopReaderNormaliser(r io.Reader) io.Reader { return r }

// Match loads the output snapshot for a particular test case.  By
// default, this looks for the file at the location
// <test-directory>/__snapshots__/<test-name>/output.txt. The base directory
// cannot be changed, however the file name can be overridden using with the
// relevant MatchOption arguments.  If the output snapshot file does not
// exist, the input actual is used in its place and the test is likely to
// succeed. In this case actual is also persisted to the disk for use in
// subsequent test runs.
func Match(t *testing.T, actual io.Reader, optFns ...MatchOption) (ok bool, msg string) {
	opts := MatchOptions{
		SnapshotName:     "output",
		FileExtension:    ".txt",
		Comparator:       StringComparator,
		ReaderNormaliser: NopReaderNormaliser,
	}
	for _, optFn := range optFns {
		optFn(&opts)
	}
	p := filepath.Clean(getSnapshotFilePath(t, opts.SnapshotName, opts.FileExtension))
	t.Logf("output snapshot filename: %v", p)
	var expected io.Reader
	if file, err := os.Open(p); err == nil {
		t.Logf("using existing snapshot")
		expected = file
		t.Cleanup(func() { _ = file.Close() })
	} else if os.IsNotExist(err) {
		t.Log("creating new output snapshot")
		err = os.MkdirAll(filepath.Dir(p), 0750)
		if err != nil {
			t.Fatalf("failed to create output snapshot file %v: %v", p, err.Error())
		}
		file, err := os.Create(p)
		if err != nil {
			t.Fatalf("failed to open newly created snapshot file: %v: %v", p, err.Error())
		}
		t.Cleanup(func() { _ = file.Close() })
		actualCopy := new(bytes.Buffer)
		_, err = io.Copy(io.MultiWriter(file, actualCopy), actual)
		if err != nil {
			t.Fatalf("failed to write to newly created snapshot file: %v: %v", p, err.Error())
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			t.Fatalf("failed to seek to beginning for snapshot file: %v", err.Error())
		}
		expected = file
		actual = actualCopy
	}
	ok, msg = opts.Comparator(
		opts.ReaderNormaliser(expected),
		opts.ReaderNormaliser(actual),
	)
	return
}
