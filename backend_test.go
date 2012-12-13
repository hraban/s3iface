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

package s3iface

import (
	"fmt"
	"testing"

	goamzs3 "launchpad.net/goamz/s3"
)

func testBucketSimple(b Bucket, errs chan<- error) {
	var err error
	var testname string
	defer func() {
		if err != nil {
			errs <- fmt.Errorf("Failed bucket test (simple) %q: %v", testname, err)
		}
	}()
	testname = "Put test.txt"
	err = b.Put("test.txt", []byte("Hello!"), "text/plain", goamzs3.Private)
	if err != nil {
		return
	}
	testname = "List objects"
	ls, err := b.List("", "/", "", 0)
	if err != nil {
		return
	}
	if len(ls.Contents) != 1 || ls.Contents[0].Key != "test.txt" {
		err = fmt.Errorf("Unexpected contents: %+v", ls.Contents)
		return
	}
	testname = "Get test.txt"
	data, err := b.Get("test.txt")
	if err != nil {
		return
	}
	if string(data) != "Hello!" {
		err = fmt.Errorf("Got wrong test data back: %q", data)
		return
	}
	testname = "Del test.txt"
	err = b.Del("test.txt")
	return
}

func testBucketNested(b Bucket, errs chan<- error) {
	var err error
	var testname string
	defer func() {
		if err != nil {
			errs <- fmt.Errorf("Failed bucket test (nested) %q: %v", testname, err)
		}
	}()
	testname = "Put nested objects"
	err = b.Put("a/b/test.txt", []byte("Hello!"), "text/plain", goamzs3.Private)
	if err != nil {
		return
	}
	err = b.Put("a/foo/bar.txt", []byte("NananananaBATMAN!"), "text/plain", goamzs3.Private)
	if err != nil {
		return
	}
	err = b.Put("a/imina", []byte("nobodylikesq"), "text/plain", goamzs3.Private)
	if err != nil {
		return
	}
	testname = "List objects"
	ls, err := b.List("a/", "/", "", 0)
	if err != nil {
		return
	}
	if len(ls.Contents) != 1 || ls.Contents[0].Key != "a/imina" {
		err = fmt.Errorf("Unexpected contents: %+v", ls.Contents)
		return
	}
	if len(ls.CommonPrefixes) != 2 || ls.CommonPrefixes[0] != "a/b/" || ls.CommonPrefixes[1] != "a/foo/" {
		err = fmt.Errorf("Unexpected common prefixes: %+v", ls.CommonPrefixes)
		return
	}
	testname = "Get a/b/test.txt"
	data, err := b.Get("a/b/test.txt")
	if err != nil {
		return
	}
	if string(data) != "Hello!" {
		err = fmt.Errorf("Got wrong test data back: %q", data)
		return
	}
	testname = "Del a/b/test.txt"
	err = b.Del("a/b/test.txt")
	return
}

// Delete the bucket completely (think rm -rf ...)
func purgeBucket(b Bucket, errs chan<- error) {
	defer func() {
		errs <- b.DelBucket()
	}()
	ls, err := b.List("", "", "", 0)
	if err != nil {
		errs <- fmt.Errorf("Failed to list bucket for cleanup: %v", err)
		return
	}
	for _, key := range ls.Contents {
		errs <- b.Del(key.Key)
	}
	return
}

func runS3tests(s3 S3, errs chan<- error) {
	b := s3.Bucket("test")
	err := b.PutBucket(goamzs3.Private)
	if err != nil {
		errs <- fmt.Errorf("Failed to create test bucket: %v", err)
	}
	// Try to clean up, do not care if it fails
	defer purgeBucket(b, errs)
	testBucketSimple(b, errs)
	testBucketNested(b, errs)
	return
}

func testS3backend(backend S3, t *testing.T) {
	errs := make(chan error)
	go func() {
		runS3tests(backend, errs)
		close(errs)
	}()
	for err := range errs {
		if err != nil {
			t.Error("S3 backend:", err)
		}
	}
}
