package beater

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// loggingRoundTripper provides a way to dump HTTP requests/responses to stdout.
// WARNING: this will include credentials so don't use it in production!
type loggingRoundTripper struct {
	Transport http.RoundTripper
}

func (rt loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if requestText, derr := httputil.DumpRequestOut(req, true); derr == nil {
		log.Println("HTTP Request:", string(requestText))
	}

	log.Println("Executing HTTP Request")
	resp, err := rt.Transport.RoundTrip(req)

	if responseText, derr := httputil.DumpResponse(resp, true); derr == nil {
		log.Println("Got HTTP Response:", string(responseText))
	}

	return resp, err
}

func LogClientHttpRequests(client *http.Client) {
	parent := client.Transport

	if parent == nil {
		parent = http.DefaultTransport
	}

	log.Println("Client transport:", parent)
	rt := loggingRoundTripper{parent}
	client.Transport = rt
}
