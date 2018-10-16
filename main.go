package main

import (
	"os"

	"github.com/imduffy15/aws-keycloak-cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// VERSION defines the cli version
var VERSION = "v0.0.0-dev"

var appHelpTemplate = `{{.Usage}}

Usage: {{.Name}} {{if .Flags}}[GLOBAL_OPTIONS] {{end}}COMMAND [arg...]

Version: {{.Version}}
{{if .Flags}}
Options:
  {{range .Flags}}{{if .Hidden}}{{else}}{{.}}
  {{end}}{{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

var commandHelpTemplate = `{{.Usage}}
{{if .Description}}{{.Description}}{{end}}
Usage: assume-role-with-keycloak [global options] {{.Name}} {{if .Flags}}[OPTIONS] {{end}}{{if ne "None" .ArgsUsage}}{{if ne "" .ArgsUsage}}{{.ArgsUsage}}{{else}}[arg...]{{end}}{{end}}

{{if .Flags}}Options:{{range .Flags}}
	 {{.}}{{end}}{{end}}
`

func main() {
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate

	app := cli.NewApp()
	app.Name = "assume-role-with-keycloak"
	app.Version = VERSION
	app.Usage = "CLI tool for executing AWS AssumeRoleWithSAML using keycloak"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		logrus.Debugf("assume-role-with-keycloak version: %v", VERSION)
		return nil
	}
	app.Author = "Ian Duffy"
	app.Commands = []cli.Command{
		cmd.LoginCommand(),
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable verbose logging",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
