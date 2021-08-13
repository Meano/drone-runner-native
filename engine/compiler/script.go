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
	dst.Envs["SHELL"] = "pwsh"
	dst.Envs["DRONE_SCRIPT"] = powershell.Script(src.Commands)
	dst.ShellEntrypoint = dst.Envs["SHELL"]
	dst.Command = []string{"-nop", "-noni", "-c", "Invoke-Expression", "$Env:DRONE_SCRIPT"}
}

// helper function configures the pipeline script for the
// linux operating system.
func setupScriptPosix(src *resource.Step, dst *engine.Step) {
	dst.Envs["SHELL"] = "/bin/sh"
	dst.Envs["DRONE_SCRIPT"] = bash.Script(src.Commands)
	dst.ShellEntrypoint = dst.Envs["SHELL"]
	dst.Command = []string{"-c", "eval $DRONE_SCRIPT"}
}
