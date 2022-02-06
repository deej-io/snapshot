package snapshot

import (
	"io"
	"path/filepath"
	"strings"
)

type SnapshotOption interface {
	GetTestInputOption
	MatchOption
}

// WithInputSnapshotName overrides the snapshot name. This is useful in cases
// where there may be multiple input/output snapshots for a test. This option does not
// change the file extension.
func WithSnapshotName(name string) SnapshotOption {
	return withSnapshotName{name}
}

type withSnapshotName struct {
	name string
}

func (wo withSnapshotName) ApplyInputOption(o *GetTestInputOptions) {
	o.SnapshotName = wo.name
}

func (wo withSnapshotName) ApplyMatchOption(o *MatchOptions) {
	o.SnapshotName = wo.name
}

// WithSnapshotFileExtension overrides the file extension for the resulting
// snapshot file. This is useful if the snapshots are read by external tools
// and make use of the extensions to determine the filetype.
func WithSnapshotFileExtension(name string) SnapshotOption {
	return withSnapshotFileExtension{name}
}

type withSnapshotFileExtension struct {
	ext string
}

func (wo withSnapshotFileExtension) ApplyInputOption(o *GetTestInputOptions) {
	o.FileExtension = wo.ext
}

func (wo withSnapshotFileExtension) ApplyMatchOption(o *MatchOptions) {
	o.FileExtension = wo.ext
}

// WithSnapshotFilename overrides snapshot name and file extension of the
// resulting snapshot file.
func WithSnapshotFilename(name string) SnapshotOption {
	return withSnapshotFilename{name}
}

type withSnapshotFilename struct {
	filename string
}

func (wo withSnapshotFilename) ApplyInputOption(o *GetTestInputOptions) {
	o.FileExtension = filepath.Ext(wo.filename)
	o.SnapshotName = strings.TrimSuffix(wo.filename, o.FileExtension)
}

func (wo withSnapshotFilename) ApplyMatchOption(o *MatchOptions) {
	o.FileExtension = filepath.Ext(wo.filename)
	o.SnapshotName = strings.TrimSuffix(wo.filename, o.FileExtension)
}

// WithCreateSnapshot provides a SnapshotCreator function to specify the
// data for the test when no snapshot file exists. This data is also persisted
// to disk and used for subsequent test runs.
func WithCreateSnapshot(r func() (io.Reader, error)) GetTestInputOption {
	return GetTestInputOptionFunc(func(o *GetTestInputOptions) { o.CreateSnapshot = r })
}

// GetTestInputOptionFunc applies a func to the GetTestInputOptions defaults.
type GetTestInputOptionFunc func(*GetTestInputOptions)

func (giof GetTestInputOptionFunc) ApplyInputOption(o *GetTestInputOptions) {
	giof(o)
}

// WithCreateSnapshotFromReader creates SnapshotCreator func from a reader.
func WithCreateSnapshotFromReader(r io.Reader) GetTestInputOption {
	return WithCreateSnapshot(func() (io.Reader, error) { return r, nil })
}

// WithComparator overrides the default comparator with a custom function.
func WithComparator(cmp Comparator) MatchOption {
	return MatchOptionFunc(func(o *MatchOptions) { o.Comparator = cmp })
}

// WithReaderNormaliser provides a ReaderNormaliser function to apply to the
// actual and expected io.Readers before comparison.
func WithReaderNormaliser(rn ReaderNormaliser) MatchOption {
	return MatchOptionFunc(func(o *MatchOptions) { o.ReaderNormaliser = rn })
}

// WithMatchOption applies f to the MatchOptions defaults.
type MatchOptionFunc func(*MatchOptions)

func (mof MatchOptionFunc) ApplyMatchOption(o *MatchOptions) {
	mof(o)
}
