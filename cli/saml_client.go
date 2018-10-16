package cli

import (
	"encoding/base64"
	"net/url"

	"github.com/sirupsen/logrus"
)

type BlehBleh interface {
	Start()
	Authorize()
	Done() chan []byte
}

type SamlClient struct {
	URL                string
	Port               int
	SamlCallbackServer SamlCallbackServer
	BrowserLauncher    func(string) error
	done               chan []byte
}

const CallbackCSS = `<style>
	@import url('https://fonts.googleapis.com/css?family=Source+Sans+Pro');
	html {
		background: #f8f8f8;
		font-family: "Source Sans Pro", sans-serif;
	}
</style>`
const CallbackJS = ``
const CallbackHTML = `<body>
	<h1>Authentication: Success</h1>
	<p>Keycloak redirected you to this page with a SAML Assertion.</p>
	<p> The SAML Assertion will be used to execute AWS's AssumeRoleWithSAML. You may close this window.</p>
</body>`

func NewSamlClient(keycloakURL string, port int, launcher func(string) error) SamlClient {
	samlClient := SamlClient{
		URL:             keycloakURL,
		Port:            port,
		BrowserLauncher: launcher,
		done:            make(chan []byte),
	}

	callbackServer := NewSamlCallbackServer(CallbackHTML, CallbackCSS, CallbackJS, port)
	callbackServer.SetHangupFunc(func(done chan url.Values, values url.Values) {
		samlAssertion := values.Get("SAMLResponse")
		if samlAssertion != "" {
			done <- values
		}
	})
	samlClient.SamlCallbackServer = callbackServer

	return samlClient
}

func (samlClient SamlClient) Start() {
	go func() {
		urlValues := make(chan url.Values)
		go samlClient.SamlCallbackServer.Start(urlValues)
		values := <-urlValues

		samlAssertion, err := base64.StdEncoding.DecodeString(values.Get("SAMLResponse"))
		if err != nil {
			logrus.Fatalf("Failed to decode the SAML Assertion", err)
		}

		samlClient.Done() <- samlAssertion
	}()
}
func (samlClient SamlClient) Authorize() {
	logrus.Info("Launching browser window to " + samlClient.URL)
	samlClient.BrowserLauncher(samlClient.URL)
}
func (samlClient SamlClient) Done() chan []byte {
	return samlClient.done
}
