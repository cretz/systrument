package remote
import (
	"github.com/cretz/systrument/context"
	"fmt"
	"errors"
	"github.com/cretz/systrument/util"
)

type Remote struct {
	Server *RemoteServer `json:"server"`
	ctx *context.Context
}

type RemoteServer struct {
	Host string `json:"host"`
	SSH *SSH `json:"ssh"`
}

type SSH struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Sudo bool `json:"sudo"`
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
		errs = append(errs, errors.New("Remote sserver 'ssh' required"))
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
	return fmt.Errorf("NO!!")
}