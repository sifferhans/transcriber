// Package procutil holds small helpers for spawning subprocesses safely.
package procutil

import (
	"os/exec"
	"syscall"
	"time"
)

// KillGroupOnCancel makes a context cancel SIGKILL the child's entire process
// group, not just the parent. WaitDelay forces pipes closed if any grandchild
// keeps them open — without it Cmd.Wait can block indefinitely after Cancel.
func KillGroupOnCancel(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = 5 * time.Second
}
