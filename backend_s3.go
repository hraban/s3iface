package s3iface

// Cast a real S3 object to the S3 interface

import (
	goamzs3 "launchpad.net/goamz/s3"
)

type s3alias goamzs3.S3

func (s *s3alias) Bucket(name string) Bucket {
	return (*goamzs3.S3)(s).Bucket(name)
}

func WrapS3(s3 *goamzs3.S3) S3 {
	return (*s3alias)(s3)
}
