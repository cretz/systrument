package context

import (
	"fmt"
	"github.com/cretz/systrument/resource"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type remoteResources struct {
	ctx     *Context
	counter int
}

func newRemoteResources(ctx *Context) resource.Resources {
	return &remoteResources{ctx, 0}
}

func (r *remoteResources) ReadFile(localPath string) ([]byte, error) {
	f, err := r.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (r *remoteResources) Open(localPath string) (*os.File, error) {
	tempFile, err := r.ReadableFileName(localPath)
	if err != nil {
		return nil, err
	}
	return os.Open(tempFile)
}

func (r *remoteResources) ReadableFileName(localPath string) (string, error) {
	// TODO: atomic
	r.counter++
	tempFile := filepath.Join(r.ctx.TempDir, "temp-file-"+strconv.Itoa(r.counter))
	_, err := r.ctx.RemotePipe.Request("send-file " + localPath + " --to-- " + tempFile)
	if err != nil {
		return "", fmt.Errorf("Failed to obtain file from remote: %v", err)
	}
	return tempFile, nil
}
