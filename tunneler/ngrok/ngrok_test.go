package ngrok

import (
	"net/url"
	"strings"
	"testing"
)

func TestOpen(t *testing.T) {
	tunnel := NewTunnel(true)
	tunnelURL, err := tunnel.Open(7070)
	if err != nil {
		t.Fatal("Error opening tunnel. Recieved error: ", err)
	}
	defer tunnel.Close()
	// now that we know the tunnel is open, let's make sure the URL makes sense.
	// this really just means that the url contains the string "ngrok.io" with
	// something before and after it.
	url, err := url.Parse(tunnelURL)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(url.Host, "ngrok.io") {
		t.Error("ngrok.io not detected in returned url")
	}
}
