package errgoengine

import (
	"io"
	"io/fs"
	"os"
)

type RawFS struct{}

func (*RawFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (rfs *RawFS) ReadFile(name string) ([]byte, error) {
	file, err := rfs.Open(name)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}
