// Copyright (c) 2021 Meano

package daemon

import (
	"github.com/meano/drone-runner-native/engine"
	"github.com/meano/drone-runner-native/engine/compiler"
	"github.com/meano/drone-runner-native/engine/linter"
	"github.com/meano/drone-runner-native/engine/resource"

	"github.com/drone/runner-go/client"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/pipeline/reporter/remote"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/secret"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type processCommand struct {
	stage int64
}

func (c *processCommand) run(*kingpin.ParseContext) error {
	// load the configuration from the environment
	config, err := fromEnviron()
	if err != nil {
		return err
	}

	// setup the global logrus logger.
	setupLogger(config)

	cli := client.New(
		config.Client.Address,
		config.Client.Secret,
		config.Client.SkipVerify,
	)
	if config.Client.Dump {
		cli.Dumper = logger.StandardDumper(
			config.Client.DumpBody,
		)
	}
	cli.Logger = logger.Logrus(
		logrus.NewEntry(
			logrus.StandardLogger(),
		),
	)

	engine, err := engine.New()
	if err != nil {
		logrus.WithError(err).
			Fatalln("cannot load the docker engine")
	}

	remote := remote.New(cli)

	runner := &runtime.Runner{
		Client:   cli,
		Machine:  config.Runner.Name,
		Environ:  config.Runner.Environ,
		Reporter: remote,
		Lookup:   resource.Lookup,
		Lint:     linter.New().Lint,
		Match:    nil,
		Compiler: &compiler.Compiler{
			Clone:          config.Runner.Clone,
			NetrcCloneOnly: config.Netrc.CloneOnly,
			Resources: compiler.Resources{
				Memory:     config.Resources.Memory,
				MemorySwap: config.Resources.MemorySwap,
				CPUQuota:   config.Resources.CPUQuota,
				CPUPeriod:  config.Resources.CPUPeriod,
				CPUShares:  config.Resources.CPUShares,
				CPUSet:     config.Resources.CPUSet,
				ShmSize:    config.Resources.ShmSize,
			},
			Environ: provider.Combine(
				provider.Static(config.Runner.Environ),
				provider.External(
					config.Environ.Endpoint,
					config.Environ.Token,
					config.Environ.SkipVerify,
				),
			),
			Secret: secret.Combine(
				secret.StaticVars(
					config.Runner.Secrets,
				),
				secret.External(
					config.Secret.Endpoint,
					config.Secret.Token,
					config.Secret.SkipVerify,
				),
			),
		},
		Exec: runtime.NewExecer(
			remote,
			remote,
			engine,
			config.Runner.Procs,
		).Exec,
	}

	err = runner.RunAccepted(nocontext, c.stage)
	if err != nil {
		logrus.WithError(err).Errorln("pipeline execution failed")
	}
	return nil
}

func registerProcess(app *kingpin.Application) {
	c := new(processCommand)

	cmd := app.Command("process", "processes a pipeline by id").
		Action(c.run).
		Hidden()

	cmd.Arg("id", "pipeline id").
		Required().
		Int64Var(&c.stage)
}
