// Copyright (c) 2021 Meano

package compiler

import (
	stdpath "path"
	"strings"

	"github.com/meano/drone-runner-native/engine"
	"github.com/meano/drone-runner-native/engine/resource"
)

const (
	workspacePath     = "src"
	workspaceName     = "workspace/"
	workspaceHostName = "host"
)

var (
	root = ""
)

func createWorkspace(from *resource.Pipeline) (base, path, full string) {
	base = from.Workspace.Base
	path = from.Workspace.Path
	if base == "" {
		if strings.HasPrefix(path, "/") {
			base = path
			path = ""
		} else {
			base = workspaceName + random() + "/"
			path = workspacePath
		}
	}
	full = stdpath.Join(base, path)

	if from.Platform.OS == "windows" {
		base = toWindowsDrive(base)
		path = toWindowsPath(path)
		full = toWindowsDrive(full)
	}
	return base, path, full
}

func setupWorkdir(src *resource.Step, dst *engine.Step, path string) {
	// if the working directory is already set
	// do not alter.
	if dst.WorkingDir != "" {
		return
	}
	// if the user is running the container as a
	// service (detached mode) with no commands, we
	// should use the default working directory.
	if dst.Detach && len(src.Commands) == 0 {
		return
	}
	// else set the working directory.
	dst.WorkingDir = path
}

// helper function converts the path to a valid windows
// path, including the default C drive.
func toWindowsDrive(s string) string {
	prefix := "C:"
	if root != "" {
		prefix = root
	}
	return prefix + toWindowsPath(s)
}

// helper function converts the path to a valid windows
// path, replacing backslashes with forward slashes.
func toWindowsPath(s string) string {
	return strings.Replace(s, "/", "\\", -1)
}
