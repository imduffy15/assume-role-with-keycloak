package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

type SamlCallbackServer struct {
	html       string
	css        string
	javascript string
	port       int
	hangupFunc func(chan url.Values, url.Values)
}

func NewSamlCallbackServer(html, css, js string, port int) SamlCallbackServer {
	server := SamlCallbackServer{html: html, css: css, javascript: js, port: port}
	server.SetHangupFunc(func(done chan url.Values, vals url.Values) {})
	return server
}

func (samlCallbackServer SamlCallbackServer) Html() string {
	return samlCallbackServer.html
}
func (samlCallbackServer SamlCallbackServer) CSS() string {
	return samlCallbackServer.css
}
func (samlCallbackServer SamlCallbackServer) Javascript() string {
	return samlCallbackServer.javascript
}

func (samlCallbackServer SamlCallbackServer) Port() int {
	return samlCallbackServer.port
}

func (samlCallbackServer SamlCallbackServer) Hangup(done chan url.Values, values url.Values) {
	samlCallbackServer.hangupFunc(done, values)
}
func (samlCallbackServer *SamlCallbackServer) SetHangupFunc(hangupFunc func(chan url.Values, url.Values)) {
	samlCallbackServer.hangupFunc = hangupFunc
}

func (samlCallbackServer SamlCallbackServer) Start(done chan url.Values) {
	callbackValues := make(chan url.Values)
	serveMux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", samlCallbackServer.port),
		Handler: serveMux,
	}

	go func() {
		value := <-callbackValues
		close(callbackValues)
		srv.Close()
		done <- value
	}()

	attemptHangup := func(formValues url.Values) {
		time.Sleep(10 * time.Millisecond)
		samlCallbackServer.Hangup(callbackValues, formValues)
	}

	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, samlCallbackServer.css)
		io.WriteString(w, samlCallbackServer.html)
		io.WriteString(w, samlCallbackServer.javascript)
		logrus.Infof("Local server received request to %v %v", r.Method, r.RequestURI)

		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		// This is a goroutine because we want this handleFunc to complete before
		// Server.Close is invoked by listeners on the callbackValues channel.
		go attemptHangup(r.Form)
	})

	go func() {
		logrus.Infof("Starting local HTTP server on port %v", samlCallbackServer.port)
		logrus.Info("Waiting for SAML Assertion from Keycloak...")
		if err := srv.ListenAndServe(); err != nil {
			logrus.Infof("Stopping local HTTP server on port %v", samlCallbackServer.port)
		}
	}()
}
