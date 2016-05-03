package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/shell"
	"github.com/cretz/systrument/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Remote struct {
	Server *RemoteServer `json:"server"`
	ctx    *context.Context
	ssh    *sshConn
}

type RemoteServer struct {
	Host string `json:"host"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
	SSH  *SSH   `json:"ssh"`
}

type SSH struct {
	User       string `json:"user"`
	Port       int    `json:"port"`
	Pass       string `json:"pass"`
	PrivateKey string `json:"privateKey"`
	Sudo       bool   `json:"sudo"`
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
		if (r.SSH.Pass == "") == (r.SSH.PrivateKey == "") {
			errs = append(errs, errors.New("Remote server 'ssh.pass' or 'ssh.privateKey' required but not both"))
		}
	}
	return
}

func (r *Remote) RunRemotely() error {
	localFile, err := r.buildForRemote()
	if err != nil {
		return err
	}
	defer os.Remove(localFile)
	r.ctx.Infof("Sending built executable to remote system")
	if ssh, err := newSshConn(r.ctx, r.Server); err != nil {
		return err
	} else {
		r.ssh = ssh
	}
	defer r.ssh.close()
	remoteTempFile := "/tmp/" + filepath.Base(localFile)
	r.ctx.Debugf("Sending local exe %v to remote path %v", localFile, remoteTempFile)
	if err := r.ssh.sendFile(localFile, remoteTempFile, 0775); err != nil {
		return err
	}
	sess, err := r.ssh.client.NewSession()
	if err != nil {
		return fmt.Errorf("Unable to create SSH session: %v", err)
	}
	defer sess.Close()

	// We set a remote request handler on the session
	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		return fmt.Errorf("Unable to obtain stdin pipe: %v", err)
	}
	sess.Stdout = context.NewRemoteToLocalPipeListener(stdinPipe, r.handleRemoteRequest)

	// Run the command with --is-remote
	newCmdPieces := []string{
		remoteTempFile,
		"--is-remote",
		"--override-local-dir",
		// TODO: escape better
		"\"" + strings.Replace(r.ctx.BaseLocalDir, "\\", "\\\\", -1) + "\"",
	}
	cmd := strings.Join(append(newCmdPieces, os.Args[1:]...), " ")
	r.ctx.Debugf("Running command on remote: %v", cmd)
	return r.ssh.runSudoCommand(sess, stdinPipe, cmd)
}

func (r *Remote) buildForRemote() (string, error) {
	f, err := ioutil.TempFile(os.TempDir(), "syst-remote-build")
	if err != nil {
		return "", fmt.Errorf("Unable to create temp file: %v", err)
	}
	if err = f.Close(); err != nil {
		return "", fmt.Errorf("Unable to perform early close of temp file: %v", err)
	}
	// TODO: reduce size with -ldflags="-s -w"?
	cmd := shell.WrapCommandOutput(r.ctx, exec.Command("go", "build", "-o", f.Name()))
	cmd.Dir = r.ctx.BaseLocalDir
	// Set the environ the same as ours except strip GOOS and GOARCH
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "GOOS=") && !strings.HasPrefix(env, "GOARCH=") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	osName := r.Server.OS
	if osName == "" {
		osName = "linux"
	}
	arch := r.Server.Arch
	if arch == "" {
		arch = "amd64"
	}
	cmd.Env = append(cmd.Env, "GOOS="+osName, "GOARCH="+arch)
	r.ctx.Infof("Building executable for remote OS %v and arch %v", osName, arch)
	if err = cmd.Run(); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("Failed to build custom remote binary: %v", err)
	}
	return f.Name(), nil
}

func (r *Remote) handleRemoteRequest(request string) (string, error) {
	if request == "get-context-data" {
		byts, err := json.Marshal(r.ctx.Data.Values)
		if err != nil {
			return "", fmt.Errorf("Unable to marshal context data: %v", err)
		}
		return string(byts), nil
	} else if strings.HasPrefix(request, "send-file ") {
		files := strings.Split(request[10:], " --to-- ")
		if len(files) != 2 {
			return "", fmt.Errorf("Malformed request: %v", request)
		}
		return "complete", r.ssh.sendFile(files[0], files[1], os.ModePerm)
	} else {
		return "", fmt.Errorf("Unrecognized request: %v", request)
	}
}
