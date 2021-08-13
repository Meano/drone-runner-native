// Copyright (c) 2021 Meano

package compiler

import (
	"context"

	"github.com/meano/drone-runner-native/engine"
	"github.com/meano/drone-runner-native/engine/resource"

	"github.com/drone/drone-yaml/yaml/compiler/image"
	"github.com/drone/runner-go/clone"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/labels"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/secret"

	"github.com/dchest/uniuri"
)

// random generator function
var random = func() string {
	return "drone-" + uniuri.NewLen(20)
}

// Privileged provides a list of plugins that execute
// with privileged capabilities in order to run Docker
// in Docker.
var Privileged = []string{
	"plugins/docker",
	"plugins/acr",
	"plugins/ecr",
	"plugins/gcr",
	"plugins/heroku",
}

// Resources defines container resource constraints. These
// constraints are per-container, not per-pipeline.
type Resources struct {
	Memory     int64
	MemorySwap int64
	CPUQuota   int64
	CPUPeriod  int64
	CPUShares  int64
	CPUSet     []string
	ShmSize    int64
}

// Tmate defines tmate settings.
type Tmate struct {
	Image          string
	Enabled        bool
	Server         string
	Port           string
	RSA            string
	ED25519        string
	AuthorizedKeys string
}

// Compiler compiles the Yaml configuration file to an
// intermediate representation optimized for simple execution.
type Compiler struct {
	// Environ provides a set of environment variables that
	// should be added to each pipeline step by default.
	Environ provider.Provider

	// Labels provides a set of labels that should be added
	// to each container by default.
	Labels map[string]string

	// NetrcCloneOnly instrucs the compiler to only inject
	// the netrc file into the clone setp.
	NetrcCloneOnly bool

	// Clone overrides the default plugin image used
	// when cloning a repository.
	Clone string

	// Resources provides global resource constraints applies to pipeline.
	Resources Resources

	// Tate provides global configration options for tmate
	// live debugging.
	Tmate Tmate

	// Secret returns a named secret value that can be injected
	// into the pipeline step.
	Secret secret.Provider

	// Root to specify the root path for windows
	Root string
}

// Compile compiles the configuration file.
func (c *Compiler) Compile(ctx context.Context, args runtime.CompilerArgs) runtime.Spec {
	pipeline := args.Pipeline.(*resource.Pipeline)
	OS := pipeline.Platform.OS

	root = c.Root

	// create the workspace paths
	base, path, full := createWorkspace(pipeline)

	// create system labels
	labels := labels.Combine(
		c.Labels,
		labels.FromRepo(args.Repo),
		labels.FromBuild(args.Build),
		labels.FromStage(args.Stage),
		labels.FromSystem(args.System),
		labels.WithTimeout(args.Repo),
	)

	spec := &engine.Spec{
		Platform: engine.Platform{
			OS:      pipeline.Platform.OS,
			Arch:    pipeline.Platform.Arch,
			Variant: pipeline.Platform.Variant,
			Version: pipeline.Platform.Version,
		},
		Root: c.Root,
		Base: base,
	}

	// list the global environment variables
	globals, _ := c.Environ.List(ctx, &provider.Request{
		Build: args.Build,
		Repo:  args.Repo,
	})

	// create the default environment variables.
	envs := environ.Combine(
		environ.OS(pipeline.Platform.OS, base),
		provider.ToMap(
			provider.FilterUnmasked(globals),
		),
		args.Build.Params,
		pipeline.Environment,
		environ.Proxy(),
		environ.System(args.System),
		environ.Repo(args.Repo),
		environ.Build(args.Build),
		environ.Stage(args.Stage),
		environ.Link(args.Repo, args.Build, args.System),
		clone.Environ(clone.Config{
			SkipVerify: pipeline.Clone.SkipVerify,
			Trace:      pipeline.Clone.Trace,
			User: clone.User{
				Name:  args.Build.AuthorName,
				Email: args.Build.AuthorEmail,
			},
		}),
	)

	// create the workspace variables
	envs["DRONE_WORKSPACE"] = full
	envs["DRONE_WORKSPACE_BASE"] = base
	envs["DRONE_WORKSPACE_PATH"] = path

	// create tmate variables
	if c.Tmate.Server != "" {
		envs["DRONE_TMATE_HOST"] = c.Tmate.Server
		envs["DRONE_TMATE_PORT"] = c.Tmate.Port
		envs["DRONE_TMATE_FINGERPRINT_RSA"] = c.Tmate.RSA
		envs["DRONE_TMATE_FINGERPRINT_ED25519"] = c.Tmate.ED25519

		if c.Tmate.AuthorizedKeys != "" {
			envs["DRONE_TMATE_AUTHORIZED_KEYS"] = c.Tmate.AuthorizedKeys
		}
	}

	// create the .netrc environment variables if not
	// explicitly disabled
	if !c.NetrcCloneOnly {
		envs = environ.Combine(envs, environ.Netrc(args.Netrc))
	}

	match := manifest.Match{
		Action:   args.Build.Action,
		Cron:     args.Build.Cron,
		Ref:      args.Build.Ref,
		Repo:     args.Repo.Slug,
		Instance: args.System.Host,
		Target:   args.Build.Deploy,
		Event:    args.Build.Event,
		Branch:   args.Build.Target,
	}

	// create the clone step
	if !pipeline.Clone.Disable {
		step := createClone(pipeline)
		step.ID = random()
		step.Envs = environ.Combine(envs, step.Envs)
		step.WorkingDir = full
		step.Labels = labels
		step.Pull = engine.PullIfNotExists
		// step.Volumes = append(step.Volumes, mount)
		spec.Steps = append(spec.Steps, step)

		// always set the .netrc file for the clone step.
		// note that environment variables are only set
		// if the .netrc file is not nil (it will always
		// be nil for public repositories).
		step.Envs = environ.Combine(step.Envs, environ.Netrc(args.Netrc))

		// if the clone image is customized, override
		// the default image.
		if c.Clone != "" {
			step.Image = c.Clone
		}
	}

	// create steps
	for _, src := range pipeline.Services {
		dst := createStep(pipeline, src)
		dst.Detach = true
		dst.Envs = environ.Combine(envs, dst.Envs)
		// dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = labels
		setupScript(src, dst, OS)
		setupWorkdir(src, dst, full)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = runtime.RunNever
		}
	}

	// create steps
	for _, src := range pipeline.Steps {
		dst := createStep(pipeline, src)
		dst.Envs = environ.Combine(envs, dst.Envs)
		// dst.Volumes = append(dst.Volumes, mount)
		dst.Labels = labels
		setupScript(src, dst, OS)
		setupWorkdir(src, dst, full)
		spec.Steps = append(spec.Steps, dst)

		// if the pipeline step has unmet conditions the step is
		// automatically skipped.
		if !src.When.Match(match) {
			dst.RunPolicy = runtime.RunNever
		}
	}

	// create internal steps if build running in debug mode
	if c.Tmate.Enabled && args.Build.Debug && pipeline.Platform.OS != "windows" {
		spec.Internal = append(spec.Internal, &engine.Step{
			ID:              random(),
			Labels:          labels,
			Pull:            engine.PullIfNotExists,
			Image:           image.Expand(c.Tmate.Image),
			ShellEntrypoint: "drone-runner",
			Command:         []string{"copy"},
		})
	}

	if !isGraph(spec) {
		configureSerial(spec)
	} else if !pipeline.Clone.Disable {
		configureCloneDeps(spec)
	} else if pipeline.Clone.Disable {
		removeCloneDeps(spec)
	}

	for _, step := range spec.Steps {
		for _, s := range step.Secrets {
			secret, ok := c.findSecret(ctx, args, s.Name)
			if ok {
				s.Data = []byte(secret)
			}
		}
	}

	// HACK: append masked global variables to secrets
	// this ensures the environment variable values are
	// masked when printed to the console.
	masked := provider.FilterMasked(globals)
	for _, step := range spec.Steps {
		for _, g := range masked {
			step.Secrets = append(step.Secrets, &engine.Secret{
				Name: g.Name,
				Data: []byte(g.Data),
				Mask: g.Mask,
				Env:  g.Name,
			})
		}
	}

	return spec
}

// helper function attempts to find and return the named secret from the secret provider.
func (c *Compiler) findSecret(ctx context.Context, args runtime.CompilerArgs, name string) (s string, ok bool) {
	if name == "" {
		return
	}

	// source secrets from the global secret provider
	// and the repository secret provider.
	provider := secret.Combine(
		args.Secret,
		c.Secret,
	)

	// TODO return an error to the caller if the provider
	// returns an error.
	found, _ := provider.Find(ctx, &secret.Request{
		Name:  name,
		Build: args.Build,
		Repo:  args.Repo,
		Conf:  args.Manifest,
	})
	if found == nil {
		return
	}
	return found.Data, true
}
