package util
import (
	"io"
	"bytes"
	"time"
	"errors"
	"regexp"
	"runtime"
	"path/filepath"
)

var (
	ErrTimeout = errors.New("Timed out")
	BaseDirectory = ""
)

func init() {
	_, thisFile, _, _ := runtime.Caller(1)
	baseDir, err := filepath.Abs(filepath.Join(thisFile, "../.."))
	if err != nil {
		panic(err)
	}
	BaseDirectory = baseDir
}

type readWaitResult struct {
	byts []byte
    found bool
    err error
}

func ReaderWaitFor(r io.Reader, re *regexp.Regexp, duration time.Duration) ([]byte, bool, error) {
	res := make(chan *readWaitResult, 1)
	quit := make(chan bool, 1)
	go func() {
		out := &bytes.Buffer{}
		var err error
		found := false
		var n int
		buf := make([]byte, 1024)
		for err == nil && !found {
			select {
			case <-quit:
				break
			default:
				n, err = r.Read(buf)
				if n > 0 {
					out.Write(buf[0:n])
				}
				found = re.Match(out.Bytes())
			}
		}
		res <- &readWaitResult {out.Bytes(), found, err}
	}()
	select {
	case result := <-res:
		return result.byts, result.found, result.err
	case <-time.After(duration):
		quit <- true
		return nil, false, ErrTimeout
	}
}