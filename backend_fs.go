package s3iface

// Wrap a local directory in S3 semantics

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"

	goamzs3 "launchpad.net/goamz/s3"
)

type fsbucket struct {
	s3dir fs3
	name  string
}

func (b fsbucket) root() string {
	return string(b.s3dir) + b.name + "/"
}

func (b fsbucket) full(path string) string {
	return b.root() + path
}

func parentDirName(path string) string {
	for path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return ""
	}
	return path[:idx]
}

// Delete this directory if it is empty, and continue with its parent, etc
func purgeEmptyDirs(b fsbucket, path string) error {
	if path == "" {
		return nil
	}
	err := os.Remove(b.full(path))
	if err != nil {
		ferr := err.(*os.PathError)
		if ferr.Err == syscall.ENOTEMPTY {
			err = nil
		}
		return err
	}
	return purgeEmptyDirs(b, parentDirName(path))
}

func (b fsbucket) Del(path string) error {
	err := os.Remove(b.full(path))
	if err != nil {
		return err
	}
	return purgeEmptyDirs(b, parentDirName(path))
}

func (b fsbucket) DelBucket() error {
	return os.Remove(b.root())
}

func (b fsbucket) Get(path string) ([]byte, error) {
	return ioutil.ReadFile(b.full(path))
}

func (b fsbucket) GetReader(path string) (rc io.ReadCloser, err error) {
	return os.Open(b.full(path))
}

func fi2key(fi os.FileInfo) goamzs3.Key {
	return goamzs3.Key{
		Key:          fi.Name(),
		LastModified: fi.ModTime().UTC().Format(time.RFC3339Nano),
		Size:         fi.Size(),
	}
}

func (b fsbucket) List(prefix, delim, marker string, max int) (result *goamzs3.ListResp, err error) {
	if marker != "" {
		err = errors.New("FS backend does not support a start marker")
		return
	}
	if delim != "/" {
		err = errors.New("FS backend requires a `/' delimiter")
		return
	}
	if prefix != "" && prefix[len(prefix)-1] != '/' {
		err = errors.New("FS backend only supports prefixes ending in `/'")
		return
	}
	var ls []os.FileInfo
	hasmore := false
	d, err := os.Open(b.full(prefix))
	if err != nil {
		// Treat directories that do not exist as if they have no contents
		if err.(*os.PathError).Err != syscall.ENOENT {
			err = fmt.Errorf("Opening directory to list contents failed: %v", err)
			return
		}
	} else {
		ls, err = d.Readdir(max)
		switch err {
		case nil:
			hasmore = true
			break
		case io.EOF:
			break
		default:
			err = fmt.Errorf("Listing contents of directory failed: %v", err)
			return
		}
	}
	files := make([]goamzs3.Key, 0, len(ls))
	dirs := make([]string, 0, len(ls))
	for _, fi := range ls {
		if fi.IsDir() {
			dirs = append(dirs, fi.Name()+"/")
		} else {
			files = append(files, fi2key(fi))
		}
	}
	result = &goamzs3.ListResp{
		Name:           b.name,
		Prefix:         prefix,
		Delimiter:      delim,
		MaxKeys:        max,
		Marker:         marker,
		IsTruncated:    hasmore,
		Contents:       files,
		CommonPrefixes: dirs,
	}
	return
}

// Content-type and permissions are ignored.
func (b fsbucket) Put(path string, data []byte, contType string, perm goamzs3.ACL) error {
	fullpath := b.full(path)
	if i := strings.LastIndex(path, "/"); 0 <= i {
		err := os.MkdirAll(b.full(path[:i]), 0700)
		if err != nil {
			return fmt.Errorf("Error creating parent dirs: %v", path, err)
		}
	}
	return ioutil.WriteFile(fullpath, data, 0600)
}

// Permissions are ignored
func (b fsbucket) PutBucket(perm goamzs3.ACL) error {
	return os.Mkdir(b.root(), 0700)
}

// Content-type and permissions are ignored
func (b fsbucket) PutReader(path string, r io.Reader, length int64, contType string, perm goamzs3.ACL) error {
	f, err := os.Create(b.full(path))
	if err != nil {
		return err
	}
	// Don't care about this error, the chmod is mostly cosmetic anyway
	f.Chmod(0600)
	_, err = io.CopyN(f, r, length)
	return err
}

// Not implemented in FS wrapper
func (b fsbucket) SignedURL(path string, expires time.Time) string {
	return b.full(path)
}

func (b fsbucket) URL(path string) string {
	return b.full(path)
}

type fs3 string

func (dir fs3) Bucket(name string) Bucket {
	if name[len(name)-1] == '/' {
		name = name[:len(name)-1]
	}
	return fsbucket{dir, name}
}

// Use a directory as an S3 store. As (un)safe for concurrent use as the
// underlying filesystem.
func WrapFS(dir string) S3 {
	if dir[len(dir)-1] != '/' {
		dir = dir + "/"
	}
	return fs3(dir)
}
