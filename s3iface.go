// Copyright © 2012 Hraban Luyat <hraban@0brg.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// Interfaces for objects that can model Amazon S3. Simplifies testing code
// that depends on S3 without needing a connection to AWS. Written to
// supplement the standard S3 package, "launchpad.net/goamz/s3".
package s3iface

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	goamzs3 "launchpad.net/goamz/s3"
)

type S3 interface {
	Bucket(name string) Bucket
}

type Bucket interface {
	Del(path string) error
	DelBucket() error
	Get(path string) (data []byte, err error)
	GetReader(path string) (rc io.ReadCloser, err error)
	List(prefix, delim, marker string, max int) (result *goamzs3.ListResp, err error)
	Put(path string, data []byte, contType string, perm goamzs3.ACL) error
	PutBucket(perm goamzs3.ACL) error
	PutReader(path string, r io.Reader, length int64, contType string, perm goamzs3.ACL) error
	SignedURL(path string, expires time.Time) string
	URL(path string) string
}

type s3alias goamzs3.S3

func (s *s3alias) Bucket(name string) Bucket {
	return (*goamzs3.S3)(s).Bucket(name)
}

func WrapS3(s3 *goamzs3.S3) S3 {
	return (*s3alias)(s3)
}

type fsbucket string

func (dir fsbucket) Del(path string) error {
	return os.Remove(string(dir) + path)
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

func (dir fsbucket) List(prefix, delim, marker string, max int) (result *goamzs3.ListResp, err error) {
	err = errors.New("Listing bucket contents in FS wrapper not implemented yet")
	return
}

// Content-type and permissions are ignored.
func (dir fsbucket) Put(path string, data []byte, contType string, perm goamzs3.ACL) error {
	return ioutil.WriteFile(path, data, 0600)
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
