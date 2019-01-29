package main

import (
	"git.aemian.com/dhp/client/dhpclient/command"
	"github.com/mitchellh/cli"
	"os"
)

// Commands is the mapping of all the available Consul commands.
var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}
	Commands = map[string]cli.CommandFactory{
		"request": func() (cli.Command, error) {
			return &command.RequestCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
		"subscription": func() (cli.Command, error) {
			return &command.SubscriptionCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
		"auth": func() (cli.Command, error) {
			return &command.AuthCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
		"user": func() (cli.Command, error) {
			return &command.UserCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
		"cc": func() (cli.Command, error) {
			return &command.CustomerCare{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
	}
}
