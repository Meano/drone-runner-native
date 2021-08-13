// Copyright (c) 2021 Meano

package compiler

import (
	"strconv"

	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/meano/drone-runner-native/engine"
	"github.com/meano/drone-runner-native/engine/resource"
)

const cloneStepName = "Clone"

func cloneParams(src manifest.Clone) map[string]string {
	dst := map[string]string{}
	if depth := src.Depth; depth > 0 {
		dst["PLUGIN_DEPTH"] = strconv.Itoa(depth)
	}
	if skipVerify := src.SkipVerify; skipVerify {
		dst["GIT_SSL_NO_VERIFY"] = "true"
		dst["PLUGIN_SKIP_VERIFY"] = "true"
	}
	dst["PLUGIN_PATH"] = src.Path
	return dst
}

func createClone(src *resource.Pipeline) *engine.Step {
	srcStep := &resource.Step{
		Commands: []string{"clone"},
	}

	dst := &engine.Step{
		Name:      cloneStepName,
		Image:     "Meano/clone",
		RunPolicy: runtime.RunAlways,
		Envs:      cloneParams(src.Clone),
	}

	setupScript(srcStep, dst, src.Platform.OS)
	return dst
}
