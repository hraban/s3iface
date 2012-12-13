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
	"os"
	"syscall"
	"testing"
)

func isexisterr(err error) bool {
	if err == nil {
		return false
	}
	perr, ok := err.(*os.PathError)
	if !ok {
		return false
	}
	return perr.Err == syscall.EEXIST
}

func mktmpdir(prefix string, i int) (string, error) {
	name := fmt.Sprintf("%s/%s.%05d", os.TempDir(), prefix, i)
	err := os.Mkdir(name, 0700)
	if isexisterr(err) {
		return mktmpdir(prefix, i+1)
	}
	return name, err
}

func TestWrapFS(t *testing.T) {
	const dirprefix = "test-s3iface"
	dirname, err := mktmpdir(dirprefix, 0)
	if err != nil {
		t.Fatalf("mktmpdir(%q): %v", dirprefix, err)
	}
	backend := WrapFS(dirname)
	testS3backend(backend, t)
	return
}
