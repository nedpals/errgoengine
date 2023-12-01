package errgoengine

import (
	"io"
	"io/fs"
	"os"
	"time"
)

type MultiReadFileFS struct {
	FSs []fs.ReadFileFS
}

func (mfs *MultiReadFileFS) LastAttachedIdx() int {
	return len(mfs.FSs) - 1
}

func (mfs *MultiReadFileFS) Attach(fs fs.ReadFileFS, idx int) {
	if idx < len(mfs.FSs) && idx >= 0 {
		if mfs.FSs[idx] == fs {
			return
		}

		mfs.FSs[idx] = fs
		return
	}

	// check first if the fs is already attached
	for _, f := range mfs.FSs {
		if f == fs {
			return
		}
	}

	mfs.FSs = append(mfs.FSs, fs)
}

func (mfs *MultiReadFileFS) Open(name string) (fs.File, error) {
	for _, fs := range mfs.FSs {
		if fs == nil {
			continue
		}

		if file, err := fs.Open(name); err == nil {
			return file, nil
		}
	}
	return nil, os.ErrNotExist
}

type stubFileInfo struct {
	name string
}

func (*stubFileInfo) Name() string { return "" }

func (*stubFileInfo) Size() int64 { return 0 }

func (*stubFileInfo) Mode() fs.FileMode { return 0400 }

func (*stubFileInfo) ModTime() time.Time { return time.Now() }

func (*stubFileInfo) IsDir() bool { return false }

func (*stubFileInfo) Sys() any { return nil }

type StubFile struct {
	Name string
}

func (*StubFile) Read(bt []byte) (int, error) { return 0, io.EOF }

func (vf *StubFile) Stat() (fs.FileInfo, error) { return &stubFileInfo{vf.Name}, nil }

func (*StubFile) Close() error { return nil }

type StubFS struct {
	Files []*StubFile
}

func (vfs *StubFS) StubFile(name string) *StubFile {
	file := &StubFile{
		Name: name,
	}
	vfs.Files = append(vfs.Files, file)
	return file
}

func (vfs *StubFS) Open(name string) (fs.File, error) {
	for _, file := range vfs.Files {
		if file.Name == name {
			return file, nil
		}
	}
	return nil, os.ErrNotExist
}

func (vfs *StubFS) ReadFile(name string) ([]byte, error) {
	for _, file := range vfs.Files {
		if file.Name == name {
			return make([]byte, 0), nil
		}
	}
	return nil, os.ErrNotExist
}

func (mfs *MultiReadFileFS) ReadFile(name string) ([]byte, error) {
	for _, fs := range mfs.FSs {
		if fs == nil {
			continue
		}

		if file, err := fs.Open(name); err == nil {
			return io.ReadAll(file)
		}
	}
	return nil, os.ErrNotExist
}

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
