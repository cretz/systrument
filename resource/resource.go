package resource

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Resources interface {
	ReadFile(localPath string) ([]byte, error)
	Open(localPath string) (*os.File, error)
	// TODO: join :-(
}

func CopyResource(r Resources, localPath, remotePath string) error {
	from, err := r.Open(localPath)
	if err != nil {
		return fmt.Errorf("Unable to open path: %v", err)
	}
	defer from.Close()
	to, err := os.Create(remotePath)
	if err != nil {
		return fmt.Errorf("Unable to create new file: %v", err)
	}
	defer to.Close()
	if _, err = io.Copy(to, from); err != nil {
		return fmt.Errorf("Unable to copy file: %v", err)
	}
	return nil
}

type localResources struct {
}

func LocalResources() Resources {
	return &localResources{}
}

func (_ *localResources) ReadFile(localPath string) ([]byte, error) {
	return ioutil.ReadFile(localPath)
}

func (_ *localResources) Open(localPath string) (*os.File, error) {
	return os.Open(localPath)
}
