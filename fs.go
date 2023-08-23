package errgoengine

import (
	"io"
	"io/fs"
	"os"
)

type RootFS struct{}

func (*RootFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (rfs *RootFS) ReadFile(name string) ([]byte, error) {
	file, err := rfs.Open(name)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}
