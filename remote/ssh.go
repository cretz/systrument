package remote

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/shell"
	"github.com/cretz/systrument/util"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strconv"
	"time"
)

type sshConn struct {
	*context.Context
	server *RemoteServer
	client *ssh.Client
}

func newSshConn(ctx *context.Context, server *RemoteServer) (*sshConn, error) {
	config := &ssh.ClientConfig{
		User: server.SSH.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(server.SSH.Pass),
		},
	}
	// TODO: configurable port
	port := server.SSH.Port
	if port == 0 {
		port = 22
	}
	client, err := ssh.Dial("tcp", server.Host+":"+strconv.Itoa(port), config)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to %v over SSH: %v", server.Host, err)
	}
	return &sshConn{
		Context: ctx,
		server:  server,
		client:  client,
	}, nil
}

func (s *sshConn) close() error {
	return s.client.Close()
}

func (s *sshConn) sendFile(localPath string, remotePath string, mode os.FileMode) error {
	sf, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("Unable to initiate SFTP connection: %v", err)
	}
	defer sf.Close()
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("Unable to read file at local path %v: %v", localPath, err)
	}
	defer localFile.Close()
	remoteFile, err := sf.Create(remotePath)
	if err != nil {
		return fmt.Errorf("Unable to create file at remote path %v: %v", remotePath, err)
	}
	defer remoteFile.Close()

	// This is so slow that we need to use a progress bar in debug mode
	s.Debugf("Using SFTP to send %v to %v", localPath, remotePath)
	var reader io.Reader = localFile
	if s.DebugEnabled() {
		info, err := localFile.Stat()
		if err != nil {
			return fmt.Errorf("Unable to stat local file: %v", err)
		}
		bar := pb.New(int(info.Size())).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 50)
		bar.ShowSpeed = true
		bar.Start()
		defer bar.Finish()
		reader = bar.NewProxyReader(reader)
	}

	if _, err := io.Copy(remoteFile, reader); err != nil {
		return fmt.Errorf("Unable to copy to remote file: %v", err)
	}
	if err := sf.Chmod(remotePath, mode); err != nil {
		return fmt.Errorf("Unable to chmod remote file to %v: %v", mode, err)
	}
	return nil
}

func (s *sshConn) runSudoCommand(sess *ssh.Session, stdin io.Writer, cmd string) error {
	// Wrap the output
	if s.DebugEnabled() {
		debugOutWriter := util.NewDebugLogWriter("SSH OUT:", s.Context)
		if sess.Stdout != nil {
			sess.Stdout = io.MultiWriter(sess.Stdout, debugOutWriter)
		} else {
			sess.Stdout = debugOutWriter
		}
		debugErrWriter := util.NewDebugLogWriter("SSH ERR:", s.Context)
		if sess.Stderr != nil {
			sess.Stderr = io.MultiWriter(sess.Stderr, debugErrWriter)
		} else {
			sess.Stderr = debugErrWriter
		}
	}
	// We need a checker to enter the password
	passwordTyper := util.NewExpectListener(stdin, shell.SudoPasswordPromptMatch, s.server.SSH.Pass+"\n")

	if sess.Stdout == nil {
		sess.Stdout = passwordTyper
	} else {
		sess.Stdout = io.MultiWriter(sess.Stdout, passwordTyper)
	}
	if sess.Stderr == nil {
		sess.Stderr = passwordTyper
	} else {
		sess.Stderr = io.MultiWriter(sess.Stderr, passwordTyper)
	}
	if err := sess.Run("sudo -S " + cmd); err != nil {
		return fmt.Errorf("Error running command %v: %v", cmd, err)
	}
	return nil
}
