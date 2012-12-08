// Interfaces for objects that can model Amazon S3. Simplifies testing code
// that depends on S3 without needing a connection to AWS. Written to
// supplement the standard S3 package, "launchpad.net/goamz/s3".
package s3interface

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

type s3alias goamzs3.S3

func (s *s3alias) Bucket(name string) Bucket {
	return (*goamzs3.S3)(s).Bucket(name)
}

func S3tointerface(s3 *goamzs3.S3) S3 {
	return (*s3alias)(s3)
}
