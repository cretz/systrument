package shell

import (
	"os/exec"
	"syscall"
)

func PutInBackgroundIfLinux(cmd *exec.Cmd) {
	// To put this in the background, we have to change the process group
	//	per: https://groups.google.com/forum/#!topic/golang-nuts/shST-SDqIp4.
	//	But it doesn't work on Windows
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
