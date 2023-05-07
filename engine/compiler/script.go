// Copyright (c) 2021 Meano

package compiler

import (
	"github.com/meano/drone-runner-native/engine"
	"github.com/meano/drone-runner-native/engine/compiler/shell/bash"
	"github.com/meano/drone-runner-native/engine/compiler/shell/powershell"
	"github.com/meano/drone-runner-native/engine/resource"
)

// helper function configures the pipeline script for the
// target operating system.
func setupScript(src *resource.Step, dst *engine.Step, os string) {
	if len(src.Commands) > 0 {
		switch os {
		case "windows":
			setupScriptWindows(src, dst)
		default:
			setupScriptPosix(src, dst)
		}
	}
}

// helper function configures the pipeline script for the
// windows operating system.
func setupScriptWindows(src *resource.Step, dst *engine.Step) {
	if _, hasEnv := dst.Envs["CI_SHELL"]; !hasEnv {
		dst.Envs["CI_SHELL"] = "pwsh"
	}
	dst.Envs["CI_SCRIPT"] = powershell.Script(src.Commands)
	dst.ShellEntrypoint = dst.Envs["CI_SHELL"]
	dst.Command = []string{"-nop", "-noni", "-c", "Invoke-Expression", "$Env:CI_SCRIPT"}
}

// helper function configures the pipeline script for the
// linux operating system.
func setupScriptPosix(src *resource.Step, dst *engine.Step) {
	if _, hasEnv := dst.Envs["CI_SHELL"]; !hasEnv {
		dst.Envs["CI_SHELL"] = "/bin/bash"
	}
	dst.Envs["CI_SCRIPT"] = bash.Script(src.Commands)
	dst.ShellEntrypoint = dst.Envs["CI_SHELL"]
	dst.Command = []string{"-c", "eval $CI_SCRIPT"}
}
