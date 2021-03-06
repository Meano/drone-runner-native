// Copyright (c) 2021 Meano

package compiler

import (
	"strings"

	"github.com/meano/drone-runner-native/engine"
	"github.com/meano/drone-runner-native/engine/resource"
	"github.com/meano/drone-runner-native/internal/encoder"

	"github.com/drone/runner-go/pipeline/runtime"
)

func createStep(spec *resource.Pipeline, src *resource.Step) *engine.Step {
	dst := &engine.Step{
		ID:              random(),
		Name:            src.Name,
		Image:           src.Image, // image.Expand(src.Image),
		Command:         src.Command,
		ShellEntrypoint: src.ShellEntrypoint,
		Detach:          src.Detach,
		DependsOn:       src.DependsOn,
		Envs:            convertStaticEnv(src.Environment),
		ExtraHosts:      src.ExtraHosts,
		IgnoreStderr:    false,
		IgnoreStdout:    false,
		Pull:            convertPullPolicy(src.Pull),
		User:            src.User,
		Secrets:         convertSecretEnv(src.Environment),
		ShmSize:         int64(src.ShmSize),
		WorkingDir:      src.WorkingDir,
	}

	// set container limits
	if v := int64(src.MemLimit); v > 0 {
		dst.MemLimit = v
	}
	if v := int64(src.MemSwapLimit); v > 0 {
		dst.MemSwapLimit = v
	}

	// appends the settings variables to environment
	for key, value := range src.Settings {
		if value == nil {
			continue
		}
		// all settings are passed to the plugin env
		// variables, prefixed with PLUGIN_
		key = "PLUGIN_" + strings.ToUpper(key)

		// if the setting parameter is sources from the
		// secret we create a secret enviornment variable.
		if value.Secret != "" {
			dst.Secrets = append(dst.Secrets, &engine.Secret{
				Name: value.Secret,
				Mask: true,
				Env:  key,
			})
		} else {
			// else if the setting parameter is opaque
			// we inject as a string-encoded environment
			// variable.
			dst.Envs[key] = encoder.Encode(value.Value)
		}
	}

	// set the pipeline step run policy. steps run on
	// success by default, but may be optionally configured
	// to run on failure.
	if isRunAlways(src) {
		dst.RunPolicy = runtime.RunAlways
	} else if isRunOnFailure(src) {
		dst.RunPolicy = runtime.RunOnFailure
	}

	// set the pipeline failure policy. steps can choose
	// to ignore the failure, or fail fast.
	switch src.Failure {
	case "ignore":
		dst.ErrPolicy = runtime.ErrIgnore
	case "fast", "fast-fail", "fail-fast":
		dst.ErrPolicy = runtime.ErrFailFast
	}

	return dst
}
