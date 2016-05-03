package java

import (
	"fmt"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/shell"
)

type Java struct {
}

func (_ *Java) Install(ctx *context.Context) error {
	// Add repo
	if err := shell.Run(ctx, "add-apt-repository", "ppa:webupd8team/java"); err != nil {
		return fmt.Errorf("Unable to add repo: %v", err)
	}
	if err := shell.Run(ctx, "apt-get", "update"); err != nil {
		return fmt.Errorf("Unable to update apt: %v", err)
	}

	// Pre-accept the Oracle license and install
	cmdLine := "echo 'oracle-java8-installer shared/accepted-oracle-license-v1-1 select true' | debconf-set-selections"
	if err := shell.Run(ctx, "bash", "-c", cmdLine); err != nil {
		return fmt.Errorf("Unable to accept Oracle license: %v", err)
	}
	if err := shell.Run(ctx, "apt-get", "install", "-y", "oracle-java8-installer"); err != nil {
		return fmt.Errorf("Unable to install necessities: %v", err)
	}
	return nil
}
