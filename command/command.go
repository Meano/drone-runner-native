// Copyright (c) 2021 Meano

package command

import (
	"os"

	"github.com/meano/drone-runner-native/command/daemon"

	"gopkg.in/alecthomas/kingpin.v2"
)

var version = "0.1.0"

func Command() {
	app := kingpin.New("drone-runner", "drone native runner")
	daemon.Register(app)

	kingpin.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
