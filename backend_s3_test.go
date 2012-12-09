package s3iface

import (
	"testing"

	"launchpad.net/goamz/aws"
	goamzs3 "launchpad.net/goamz/s3"
)

func TestWrapS3(t *testing.T) {
	auth, err := aws.EnvAuth()
	if err != nil {
		t.Fatalf("Need AWS auth for testing: %v", err)
	}
	// TODO: CL switch or similar to change the AWS zone
	backend := WrapS3(goamzs3.New(auth, aws.EUWest))
	testS3backend(backend, t)
	return
}
