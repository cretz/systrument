package git

import (
	"fmt"
	"github.com/cretz/systrument/context"
	"github.com/cretz/systrument/shell"
	"github.com/cretz/systrument/util"
	"github.com/hashicorp/go-version"
	"os/exec"
	"time"
)

type Git struct {
	*context.Context
}

func NewGit(ctx *context.Context) *Git {
	return &Git{ctx}
}

func (g *Git) Version() (*version.Version, error) {
	byts, err := shell.CombinedOutput(g.Context, "git", "version")
	if err != nil {
		return nil, fmt.Errorf("Failed asking for git version: %v", err)
	}
	return util.VersionAfterLastSpace(string(byts))
}

func (g *Git) Clone(repo *Repo, intoDir string) error {
	// TODO: --single-branch?
	properUrl, err := repo.URLWithCredentials()
	if err != nil {
		return fmt.Errorf("Invalid URL: %v", err)
	}
	cmd := shell.WrapCommandOutput(g.Context, exec.Command("git", "clone", "-b", repo.Branch, properUrl, intoDir))
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("Unable to clone: %v", err)
	}
	// Wait for completion
	if err = shell.WaitTimeout(cmd, 30*time.Second); err != nil {
		return fmt.Errorf("Failed clone: %v", err)
	}
	return nil
}

func (g *Git) ResetHard(dir string) error {
	cmd := shell.WrapCommandOutput(g.Context, exec.Command("git", "reset", "--hard"))
	cmd.Dir = dir
	return cmd.Run()
}

func (g *Git) Pull(repo *Repo, dir string) error {
	properUrl, err := repo.URLWithCredentials()
	if err != nil {
		return fmt.Errorf("Invalid URL: %v", err)
	}
	cmd := shell.WrapCommandOutput(g.Context, exec.Command("git", "pull", properUrl))
	cmd.Dir = dir
	return cmd.Run()
}
