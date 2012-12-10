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

type fsbucket string

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
func purgeEmptyDirs(dir fsbucket, path string) error {
	if path == "" {
		return nil
	}
	err := os.Remove(string(dir) + path)
	if err != nil {
		ferr := err.(*os.PathError)
		if ferr.Err == syscall.ENOTEMPTY {
			err = nil
		}
		return err
	}
	return purgeEmptyDirs(dir, parentDirName(path))
}

func (dir fsbucket) Del(path string) error {
	err := os.Remove(string(dir) + path)
	if err != nil {
		return err
	}
	return purgeEmptyDirs(dir, parentDirName(path))
}

func (dir fsbucket) DelBucket() error {
	return os.Remove(string(dir))
}

func (dir fsbucket) Get(path string) ([]byte, error) {
	return ioutil.ReadFile(string(dir) + path)
}

func (dir fsbucket) GetReader(path string) (rc io.ReadCloser, err error) {
	return os.Open(string(dir) + path)
}

func fi2key(fi os.FileInfo) goamzs3.Key {
	return goamzs3.Key{
		Key:          fi.Name(),
		LastModified: fi.ModTime().UTC().Format(time.RFC3339Nano),
		Size:         fi.Size(),
	}
}

func (dir fsbucket) List(prefix, delim, marker string, max int) (result *goamzs3.ListResp, err error) {
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
	d, err := os.Open(string(dir) + prefix)
	if err != nil {
		err = fmt.Errorf("Opening directory to list contents failed: %v", err)
		return
	}
	ls, err := d.Readdir(max)
	var hasmore bool
	switch err {
	case nil:
		hasmore = true
		break
	case io.EOF:
		hasmore = false
		break
	default:
		err = fmt.Errorf("Listing contents of directory failed: %v", err)
		return
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
		// TODO: bucket name
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
func (dir fsbucket) Put(path string, data []byte, contType string, perm goamzs3.ACL) error {
	fullpath := string(dir) + path
	if i := strings.LastIndex(path, "/"); 0 <= i {
		err := os.MkdirAll(string(dir)+path[:i], 0700)
		if err != nil {
			return fmt.Errorf("Error creating parent dirs: %v", path, err)
		}
	}
	return ioutil.WriteFile(fullpath, data, 0600)
}

// Permissions are ignored
func (dir fsbucket) PutBucket(perm goamzs3.ACL) error {
	return os.Mkdir(string(dir), 0700)
}

// Content-type and permissions are ignored
func (dir fsbucket) PutReader(path string, r io.Reader, length int64, contType string, perm goamzs3.ACL) error {
	f, err := os.Create(string(dir) + path)
	if err != nil {
		return err
	}
	// Don't care about this error, the chmod is mostly cosmetic anyway
	f.Chmod(0600)
	_, err = io.CopyN(f, r, length)
	return err
}

// Not implemented in FS wrapper
func (dir fsbucket) SignedURL(path string, expires time.Time) string {
	return string(dir) + path
}

func (dir fsbucket) URL(path string) string {
	return string(dir) + path
}

type fs3 string

func (dir fs3) Bucket(name string) Bucket {
	if !strings.HasSuffix(name, "/") {
		name = name + "/"
	}
	return fsbucket(string(dir) + name)
}

// Use a directory as an S3 store. As (un)safe for concurrent use as the
// underlying filesystem.
func WrapFS(dir string) S3 {
	if !strings.HasSuffix(string(dir), "/") {
		dir = dir + "/"
	}
	return fs3(dir)
}
