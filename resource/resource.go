package resource

import (
	"io/ioutil"
	"os"
)

type Resources interface {
	ReadFile(localPath string) ([]byte, error)
	Open(localPath string) (*os.File, error)
	// TODO: join :-(
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
