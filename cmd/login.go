package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/imduffy15/aws-keycloak-cli/cli"
	"github.com/imduffy15/aws-keycloak-cli/saml"
	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	urfave "github.com/urfave/cli"
)

// LoginCommand defines the login command
func LoginCommand() urfave.Command {
	return urfave.Command{
		Name:            "login",
		Usage:           "Login to AWS using Keycloak",
		Action:          loginWrapper,
		SkipFlagParsing: true,
		Flags: []urfave.Flag{
			urfave.StringFlag{
				Name:  "URL",
				Usage: "URL of keycloak",
			},
			// urfave.IntFlag{
			// 	Name:  "Port",
			// 	Usage: "Port to bind to",
			// },
		},
	}
}

func loginWrapper(ctx *urfave.Context) error {
	debug := lookUpDebugFlag()
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	done := make(chan bool)
	url := "http://localhost/auth/realms/example/protocol/saml/clients/amazon-aws-cli"
	port := 8080

	samlClient := cli.NewSamlClient(url, port, open.Run)

	go login(done, samlClient)
	<-done

	return nil
}

func login(doneRunning chan bool, samlClient cli.BlehBleh) {
	samlClient.Start()
	samlClient.Authorize()
	samlAssertion := <-samlClient.Done()

	fmt.Printf("%v\n\n", string(samlAssertion))

	saml, err := saml.Parse(samlAssertion)
	if err != nil {
		logrus.Fatalf("Failed to parse the SAML assertion", err)
	}

	for _, attrs := range saml.Attrs {
		if attrs.Name == "https://aws.amazon.com/SAML/Attributes/Role" {
			for _, val := range attrs.Values {
				fmt.Printf("ROLE: %v\n\n", val)
				// splitVal := strings.Split(val, "/")
				// role := splitVal[len(splitVal)-1]
				// fmt.Printf("[%d] %v\n", vi, role)
			}
		}
	}

	doneRunning <- true
}

func lookUpDebugFlag() bool {
	for _, arg := range os.Args {
		if arg == "--debug" {
			return true
		}
	}
	return false
}

func flagHackLookup(flagName string) string {
	// e.g. "-d" for "--driver"
	flagPrefix := flagName[1:3]

	// TODO: Should we support -flag-name (single hyphen) syntax as well?
	for i, arg := range os.Args {
		if strings.Contains(arg, flagPrefix) {
			// format '--driver foo' or '-d foo'
			if arg == flagPrefix || arg == flagName {
				if i+1 < len(os.Args) {
					return os.Args[i+1]
				}
			}

			// format '--driver=foo' or '-d=foo'
			if strings.HasPrefix(arg, flagPrefix+"=") || strings.HasPrefix(arg, flagName+"=") {
				return strings.Split(arg, "=")[1]
			}
		}
	}

	return ""
}
