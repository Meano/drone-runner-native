// Copyright (c) 2021 Meano

package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/pipeline/runtime"
)

// Runner pipeline engine.
type Runner struct {
}

// New returns a new engine.
func New() (*Runner, error) {
	return &Runner{}, nil
}

// Ping pings the Docker daemon.
func (e *Runner) Ping(ctx context.Context) error {
	return nil
}

// Setup the pipeline environment.
func (e *Runner) Setup(ctx context.Context, specv runtime.Spec) error {
	spec := specv.(*Spec)

	return TrimExtraInfo(os.MkdirAll(spec.Base, 0777))
}

// Destroy the pipeline environment.
func (e *Runner) Destroy(ctx context.Context, specv runtime.Spec) error {
	spec := specv.(*Spec)

	return TrimExtraInfo(os.RemoveAll(spec.Base))
}

// Run runs the pipeline step.
func (e *Runner) Run(ctx context.Context, specv runtime.Spec, stepv runtime.Step, output io.Writer) (*runtime.State, error) {
	log := logger.FromContext(ctx)
	spec := specv.(*Spec)
	step := stepv.(*Step)

	err := os.MkdirAll(step.WorkingDir, 0777)
	if err != nil {
		return nil, err
	}

	if step.Image != "" {
		pluginPath := "plugins/" + step.Image
		if spec.Platform.OS == "windows" {
			if spec.Root != "" {
				pluginPath = spec.Root + pluginPath
			} else {
				pluginPath = "C:" + pluginPath
			}
			pluginPath = strings.Replace(pluginPath+";", "/", "\\", -1)
		} else {
			pluginPath = pluginPath + ":"
		}
		step.Envs["PATH"] = pluginPath + step.Envs["PATH"]
	}

	cmd := exec.Command(step.ShellEntrypoint, step.Command...)
	cmd.Env = append(cmd.Env, environ.Slice(step.Envs)...)
	cmd.Dir = step.WorkingDir
	cmd.Stdout = output
	cmd.Stderr = output

	for _, secret := range step.Secrets {
		s := fmt.Sprintf("%s=%s", secret.Env, string(secret.Data))
		cmd.Env = append(cmd.Env, s)
	}

	err = cmd.Start()

	if err != nil {
		log.Error("err: ", err)
		return &runtime.State{
			ExitCode:  255,
			Exited:    true,
			OOMKilled: false,
		}, err
	}

	log = logger.FromContext(ctx)
	log = log.WithField("process.pid", cmd.Process.Pid)
	log.Debug("process started")

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err = <-done:
	case <-ctx.Done():
		cmd.Process.Kill()

		log.Debug("process killed")
		return nil, ctx.Err()
	}

	state := &runtime.State{
		ExitCode:  0,
		Exited:    true,
		OOMKilled: false,
	}
	if err != nil {
		state.ExitCode = 255
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		state.ExitCode = exiterr.ExitCode()
	}

	log.WithField("process.exit", state.ExitCode).Debug("process finished")

	return state, err
}

func TrimExtraInfo(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	i := strings.Index(s, "extra info:")
	if i > 0 {
		s = s[:i]
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, "(0x2)")
		s = strings.TrimSpace(s)
		return errors.New(s)
	}
	return err
}
