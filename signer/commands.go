package main

import (
	"git.aemian.com/dhp/signer/command"
	"github.com/mitchellh/cli"
	"os"
)

// Commands is the mapping of all the available Consul commands.
var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}
	Commands = map[string]cli.CommandFactory{
		"sign": func() (cli.Command, error) {
			return &command.SignCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
	}
}
