// Copyright (c) 2021 Meano

package engine

import (
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/pipeline/runtime"
)

type (

	// Spec provides the pipeline spec. This provides the
	// required instructions for reproducible pipeline
	// execution.
	Spec struct {
		Platform Platform `json:"platform,omitempty"`
		Steps    []*Step  `json:"steps,omitempty"`
		Internal []*Step  `json:"internal,omitempty"`
		Root     string   `json:"root,omitempty"`
		Base     string   `json:"base,omitempty"`
	}

	// Step defines a pipeline step.
	Step struct {
		ID              string            `json:"id,omitempty"`
		Command         []string          `json:"args,omitempty"`
		CPUPeriod       int64             `json:"cpu_period,omitempty"`
		CPUQuota        int64             `json:"cpu_quota,omitempty"`
		CPUShares       int64             `json:"cpu_shares,omitempty"`
		CPUSet          []string          `json:"cpu_set,omitempty"`
		Detach          bool              `json:"detach,omitempty"`
		DependsOn       []string          `json:"depends_on,omitempty"`
		ShellEntrypoint string            `json:"shell_entrypoint,omitempty"`
		Envs            map[string]string `json:"environment,omitempty"`
		ErrPolicy       runtime.ErrPolicy `json:"err_policy,omitempty"`
		ExtraHosts      []string          `json:"extra_hosts,omitempty"`
		IgnoreStdout    bool              `json:"ignore_stderr,omitempty"`
		IgnoreStderr    bool              `json:"ignore_stdout,omitempty"`
		Image           string            `json:"image,omitempty"`
		Labels          map[string]string `json:"labels,omitempty"`
		MemSwapLimit    int64             `json:"memswap_limit,omitempty"`
		MemLimit        int64             `json:"mem_limit,omitempty"`
		Name            string            `json:"name,omitempty"`
		Pull            PullPolicy        `json:"pull,omitempty"`
		RunPolicy       runtime.RunPolicy `json:"run_policy,omitempty"`
		Secrets         []*Secret         `json:"secrets,omitempty"`
		ShmSize         int64             `json:"shm_size,omitempty"`
		User            string            `json:"user,omitempty"`
		WorkingDir      string            `json:"working_dir,omitempty"`
	}

	// Secret represents a secret variable.
	Secret struct {
		Name string `json:"name,omitempty"`
		Env  string `json:"env,omitempty"`
		Data []byte `json:"data,omitempty"`
		Mask bool   `json:"mask,omitempty"`
	}

	// Platform defines the target platform.
	Platform struct {
		OS      string `json:"os,omitempty"`
		Arch    string `json:"arch,omitempty"`
		Variant string `json:"variant,omitempty"`
		Version string `json:"version,omitempty"`
	}
)

//
// implements the Spec interface
//

func (s *Spec) StepLen() int              { return len(s.Steps) }
func (s *Spec) StepAt(i int) runtime.Step { return s.Steps[i] }

//
// implements the Secret interface
//

func (s *Secret) GetName() string  { return s.Name }
func (s *Secret) GetValue() string { return string(s.Data) }
func (s *Secret) IsMasked() bool   { return s.Mask }

//
// implements the Step interface
//

func (s *Step) GetName() string                  { return s.Name }
func (s *Step) GetDependencies() []string        { return s.DependsOn }
func (s *Step) GetEnviron() map[string]string    { return s.Envs }
func (s *Step) SetEnviron(env map[string]string) { s.Envs = env }
func (s *Step) GetErrPolicy() runtime.ErrPolicy  { return s.ErrPolicy }
func (s *Step) GetRunPolicy() runtime.RunPolicy  { return s.RunPolicy }
func (s *Step) GetSecretAt(i int) runtime.Secret { return s.Secrets[i] }
func (s *Step) GetSecretLen() int                { return len(s.Secrets) }
func (s *Step) IsDetached() bool                 { return s.Detach }
func (s *Step) GetImage() string                 { return s.Image }
func (s *Step) Clone() runtime.Step {
	dst := new(Step)
	*dst = *s
	dst.Envs = environ.Combine(s.Envs)
	return dst
}
