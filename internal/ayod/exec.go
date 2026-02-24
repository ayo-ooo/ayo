package ayod

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

// Executor handles command execution as different users.
type Executor struct {
	users *UserManager
}

// NewExecutor creates a new command executor.
func NewExecutor(users *UserManager) *Executor {
	return &Executor{users: users}
}

// Exec runs a command as the specified user.
func (e *Executor) Exec(req ExecRequest) (*ExecResponse, error) {
	if len(req.Command) == 0 {
		return nil, fmt.Errorf("command required")
	}

	// Look up user
	u, err := user.Lookup(req.User)
	if err != nil {
		return nil, fmt.Errorf("unknown user %q: %w", req.User, err)
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	// Build command
	ctx := context.Background()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, req.Command[0], req.Command[1:]...)

	// Set working directory
	if req.Cwd != "" {
		cmd.Dir = req.Cwd
	} else {
		cmd.Dir = u.HomeDir
	}

	// Set environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HOME="+u.HomeDir)
	cmd.Env = append(cmd.Env, "USER="+u.Username)
	cmd.Env = append(cmd.Env, "LOGNAME="+u.Username)
	for k, v := range req.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Set credentials to run as user
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err = cmd.Run()

	// Build response
	resp := &ExecResponse{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			resp.ExitCode = 124 // Standard timeout exit code
			resp.Stderr = resp.Stderr + "\n[timeout exceeded]"
		} else {
			return nil, fmt.Errorf("exec failed: %w", err)
		}
	}

	return resp, nil
}
