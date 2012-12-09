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
	"io"
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
