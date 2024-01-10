package errgoengine

import (
	"io"
	"io/fs"
	"os"
	"reflect"
	"time"
)

type MultiReadFileFS struct {
	FSs []fs.ReadFileFS
}

func (mfs *MultiReadFileFS) LastAttachedIdx() int {
	return len(mfs.FSs) - 1
}

func (mfs *MultiReadFileFS) AttachOrReplace(fs fs.ReadFileFS, idx int) {
	if idx < len(mfs.FSs) && idx >= 0 {
		if reflect.DeepEqual(mfs.FSs[idx], fs) {
			return
		}

		mfs.FSs[idx] = fs
		return
	}

	mfs.Attach(fs, idx)
}

func (mfs *MultiReadFileFS) Attach(instance fs.ReadFileFS, idx int) {
	if idx < len(mfs.FSs) && idx >= 0 {
		if reflect.DeepEqual(mfs.FSs[idx], instance) {
			return
		}

		mfs.FSs = append(
			append(append([]fs.ReadFileFS{}, mfs.FSs[:idx]...), instance),
			mfs.FSs[idx+1:]...)
		return
	}

	// check first if the fs is already attached
	for _, f := range mfs.FSs {
		if f == instance {
			return
		}
	}

	mfs.FSs = append(mfs.FSs, instance)
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

type virtualFileInfo struct {
	name string
}

func (f *virtualFileInfo) Name() string { return f.name }

func (*virtualFileInfo) Size() int64 { return 0 }

func (*virtualFileInfo) Mode() fs.FileMode { return 0400 }

func (*virtualFileInfo) ModTime() time.Time { return time.Now() }

func (*virtualFileInfo) IsDir() bool { return false }

func (*virtualFileInfo) Sys() any { return nil }

type VirtualFile struct {
	Name string
}

func (VirtualFile) Read(bt []byte) (int, error) { return 0, io.EOF }

func (vf VirtualFile) Stat() (fs.FileInfo, error) { return &virtualFileInfo{vf.Name}, nil }

func (VirtualFile) Close() error { return nil }

type VirtualFS struct {
	Files []VirtualFile
}

func (vfs *VirtualFS) StubFile(name string) VirtualFile {
	file := VirtualFile{
		Name: name,
	}
	vfs.Files = append(vfs.Files, file)
	return file
}

func (vfs *VirtualFS) Open(name string) (fs.File, error) {
	for _, file := range vfs.Files {
		if file.Name == name {
			return file, nil
		}
	}
	return nil, os.ErrNotExist
}

func (vfs *VirtualFS) ReadFile(name string) ([]byte, error) {
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
