package remote

import (
	"errors"
	"fmt"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/shell"
	"github.com/cretz/systrument/util"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Remote struct {
	Server *RemoteServer `json:"server"`
	ctx    *context.Context
}

type RemoteServer struct {
	Host string `json:"host"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
	SSH  *SSH   `json:"ssh"`
}

type SSH struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Sudo bool   `json:"sudo"`
}

func RemoteIfPresent(ctx *context.Context) (*Remote, error) {
	r := &Remote{ctx: ctx}
	if err := ctx.Data.UnmarshalJSON(&r); err != nil {
		return nil, fmt.Errorf("Unable to fetch remote info: %v", err)
	}
	if r.Server == nil {
		return nil, nil
	}
	if errs := r.Server.validate(); len(errs) > 0 {
		return nil, fmt.Errorf("Invalid remote server: %v", util.JoinErrors(errs))
	}
	return r, nil
}

func (r *RemoteServer) validate() (errs []error) {
	if r.Host == "" {
		errs = append(errs, errors.New("Remote server 'host' required"))
	}
	if r.SSH == nil {
		errs = append(errs, errors.New("Remote server 'ssh' required"))
	} else {
		if r.SSH.User == "" {
			errs = append(errs, errors.New("Remote server 'ssh.user' required"))
		}
		if r.SSH.Pass == "" {
			errs = append(errs, errors.New("Remote server 'ssh.pass' required"))
		}
	}
	return
}

func (r *Remote) RunRemotely() error {
	f, err := ioutil.TempFile(os.TempDir(), "syst-remote-build")
	if err != nil {
		return fmt.Errorf("Unable to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if err = f.Close(); err != nil {
		return fmt.Errorf("Unable to perform early closse of temp file: %v", err)
	}
	cmd := shell.WrapCommandOutput(r.ctx, exec.Command("go", "build", "-o", f.Name()))
	cmd.Dir = r.ctx.BaseLocalDir
	// Set the environ the same as ours except strip GOOS and GOARCH
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "GOOS=") && !strings.HasPrefix(env, "GOARCH=") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	os := r.Server.OS
	if os == "" {
		os = "linux"
	}
	arch := r.Server.Arch
	if arch == "" {
		arch = "amd64"
	}
	cmd.Env = append(cmd.Env, "GOOS="+os, "GOARCH="+arch)
	r.ctx.Infof("Building executable for remote OS %v and arch %v", os, arch)
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Failed to build custom remote binary: %v", err)
	}
	return fmt.Errorf("TODO: more")
}
