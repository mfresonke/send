package ngrok

import (
	"net/url"
	"strings"
	"testing"
)

const (
	testingPort    = 7070
	testingVerbose = false
)

func TestHTTP(t *testing.T) {
	tunnel := NewHTTPTunnel(testingVerbose)
	endpoints, err := tunnel.Open(testingPort)
	if err != nil {
		t.Fatal("Error opening tunnel. Recieved error: ", err)
	}
	defer tunnel.Close()
	// Now that we know the tunnel is open, let's make sure the Endpoints make
	//  sense. Ngrok will return exactly two: one secure (https) and one insecure.
	foundSecure := false
	foundInsecure := false
	for _, ep := range endpoints {
		// Let's validate the URL itself.
		url, err := url.Parse(ep.URL)
		if err != nil {
			t.Error(err)
		}
		if !strings.Contains(url.Host, "ngrok.io") {
			t.Error("ngrok.io not detected in returned url")
		}
		isHTTPS := (url.Scheme == "https")
		// check the "Secure" flag on the endpoint struct.
		if isHTTPS != ep.Secure {
			t.Error("Secure flag on endpoint not marked properly.")
		}
		if isHTTPS && ep.Secure && !foundSecure {
			foundSecure = true
		}
		if !isHTTPS && !ep.Secure && !foundInsecure {
			foundInsecure = true
		}
	}
	if !foundSecure {
		t.Error("Did not find secure endpoint")
	}
	if !foundInsecure {
		t.Error("Did not find insecure endpoint")
	}
}

// TestDoubleClose tests if calling the "Close" method twice returns an error.
func TestDoubleClose(t *testing.T) {
	tunnel := NewHTTPTunnel(testingVerbose)
	_, err := tunnel.Open(testingPort)
	if err != nil {
		t.Fatal("Error opening tunnel. Recieved error: ", err)
	}
	err = tunnel.Close()
	if err != nil {
		t.Fatal("Error closing tunnel. Recieved error: ", err)
	}
	err = tunnel.Close()
	if err != nil {
		t.Fatal("Error closing tunnel second time. Recieved error: ", err)
	}
}

// TestTwoOpen tests if opening two ngrok connections returns an Err
func TestTwoOpen(t *testing.T) {
	tunnel1 := NewHTTPTunnel(testingVerbose)
	tunnel2 := NewHTTPTunnel(testingVerbose)
	_, err := tunnel1.Open(testingPort)
	if err != nil {
		t.Fatal("Error opening fist tunnel. Recieved error: ", err)
	}
	defer tunnel1.Close()
	_, err = tunnel2.Open(testingPort + 1)
	if err == nil {
		t.Fatal("No error returned upon opening second tunnel.")
	}
	defer tunnel2.Close()
	if err != ErrTooManyConnections {
		t.Error("ErrTooManyConnections not returned upon opening a second ngrok session")
	}
}
