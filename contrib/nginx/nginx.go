package nginx

import (
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/shell"
)

type Nginx struct {
	*context.Context
}

func NewNginx(ctx *context.Context) *Nginx {
	return &Nginx{ctx}
}

func (n *Nginx) Reload() error {
	return shell.Run(n.Context, "service", "nginx", "reload")
}

func (n *Nginx) Stop() error {
	return shell.Run(n.Context, "service", "nginx", "stop")
}

func (n *Nginx) Start() error {
	return shell.Run(n.Context, "service", "nginx", "start")
}
