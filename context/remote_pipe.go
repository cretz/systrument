package context

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// I acknowledge this std pipe communication is synchronous and naive (for now)

var requestMatcher = regexp.MustCompile("\\[syst-request-(\\d*)\\](.*)\\n")

type LocalToRemotePipe struct {
	stdin   io.Reader
	stdout  io.Writer
	counter int
}

func NewLocalToRemotePipe(stdin io.Reader, stdout io.Writer) *LocalToRemotePipe {
	return &LocalToRemotePipe{
		stdin:   stdin,
		stdout:  stdout,
		counter: 0,
	}
}

func (l *LocalToRemotePipe) Request(request string) (string, error) {
	if strings.Contains(request, "\n") {
		return "", errors.New("Request contains a newline")
	}
	l.counter++
	id := strconv.Itoa(l.counter)
	if _, err := l.stdout.Write([]byte("[syst-request-" + id + "]" + request + "\n")); err != nil {
		return "", fmt.Errorf("Unable to write request to stdout: %v", err)
	}
	// TODO: timeout please
	// Read and add to the buffer until we see the end
	buf := &bytes.Buffer{}
	tempBuf := make([]byte, 1024)
	expectedBegin := []byte("[syst-response-begin-" + id + "]\n")
	expectedEnd := []byte("\n[syst-response-end-" + id + "]\n")
	for !bytes.Contains(buf.Bytes(), expectedEnd) {
		n, err := l.stdin.Read(tempBuf)
		if err != nil {
			return "", fmt.Errorf("Error reading response: %v", err)
		}
		if _, err = buf.Write(tempBuf[0:n]); err != nil {
			return "", fmt.Errorf("Unable to fill write buffer with response: %v", err)
		}
	}
	// Now, just take the inside
	begin := bytes.Index(buf.Bytes(), expectedBegin)
	if begin == -1 {
		return "", errors.New("Beginning of response not found")
	}
	begin += len(expectedBegin)
	end := bytes.Index(buf.Bytes(), expectedEnd)
	str := string(buf.Bytes()[begin:end])
	if strings.HasPrefix(str, "ERROR: ") {
		return "", errors.New(str[7:])
	}
	return str, nil
}

type RemoteToLocalPipeListener struct {
	remoteStdin io.Writer
	handler     func(request string) (string, error)
}

func NewRemoteToLocalPipeListener(remoteStdin io.Writer, handler func(request string) (string, error)) *RemoteToLocalPipeListener {
	return &RemoteToLocalPipeListener{
		remoteStdin: remoteStdin,
		handler:     handler,
	}
}

func (r *RemoteToLocalPipeListener) Write(p []byte) (int, error) {
	matches := requestMatcher.FindSubmatch(p)
	if matches == nil || len(matches) != 3 {
		return len(p), nil
	}
	id := string(matches[1])
	resp := "[syst-response-begin-" + id + "]\n"
	str, err := r.handler(string(matches[2]))
	if err != nil {
		resp += "ERROR: " + err.Error() + "\n"
	} else {
		resp += str + "\n"
	}
	resp += "[syst-response-end-" + id + "]\n"
	if _, err := r.remoteStdin.Write([]byte(resp)); err != nil && err != io.EOF {
		return len(p), fmt.Errorf("Failure writing remote response: %v", err)
	}
	return len(p), nil
}
