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
