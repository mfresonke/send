package ngrok

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/mfresonke/send2phone/tunneler"
)

const (
	// how long to wait for ngrok before giving up
	connectionTimeout = 20 * time.Second
	// how long to wait before beginning to poll for a connection.
	initalConnectionWait = 5 * time.Second
	urlPollHTTPTimeout   = 500 * time.Millisecond
	// how long to wait between poll attempts (given the previous one did not succeed)
	urlPollDuration = 1 * time.Second
)

const tunnelsURL = "http://127.0.0.1:4040/api/tunnels"

type tunnel struct {
	*exec.Cmd
	verbose bool
}

// NewTunnel creates a new ngrok tunnel, ready to open!
func NewTunnel(verbose bool) tunneler.Interface {
	return tunnel{
		verbose: verbose,
	}
}

func (tun tunnel) Open(port int) (url string, err error) {
	if tun.verbose {
		fmt.Println("Searching for ngrok in path...")
	}
	ngrokLoc, err := exec.LookPath("ngrok")
	if err != nil {
		return "", err
	}
	if tun.verbose {
		fmt.Println("ngrok found at ", ngrokLoc)
	}

	// create and start ngrok!
	tun.Cmd = exec.Command("ngrok", "http", strconv.Itoa(port))
	err = tun.Start()
	if err != nil {
		return "", err
	}

	// channel will send an error if the process exits unexpectedly.
	errorChan := make(chan error, 1)
	go func(errorChan chan error) {
		errorChan <- tun.Wait()
	}(errorChan)

	// channel will recieve the string of the connection URL.
	waitForConnectionChan := make(chan string, 1)
	go connectionWaiter(waitForConnectionChan, tun.verbose)

	// and finally, make a channel that will time out if all else fails.
	timeoutChan := time.After(connectionTimeout)

	// wait for something to happen...
	select {
	case url := <-waitForConnectionChan:
		log.Println("Connection established: ", url)
	case err := <-errorChan:
		log.Println("An error occurred: ", err)
	case <-timeoutChan:
		log.Println("Operation timed out.")
	}
	return "", tun.Close()
}

func (tun tunnel) Close() error {
	return tun.Process.Signal(syscall.SIGTERM)
}

// connectionWaiter pings the ngrok api until it discovers a connection. Once it
// does, it sends the resulting tunnel url on the channel.
// As the channel and no return imply, this func is meant to be run asyncronously.
func connectionWaiter(waitForConnectionChan chan string, verbose bool) {
	time.Sleep(initalConnectionWait)
	firstRun := true
	for {
		if firstRun {
			firstRun = false
		} else {
			time.Sleep(urlPollDuration)
		}

		// make a request to the ngrok api to check if the connection is established.
		if verbose {
			log.Println("Making request to ngrok API to test if tunnel is online...")
		}
		resp, err := http.Get(tunnelsURL)
		if err != nil {
			if verbose {
				log.Println("Error GETing ", tunnelsURL, ", trying again...")
			}
			continue
		}
		resp.Body.Close()
		jsonDec := json.NewDecoder(resp.Body)
		var res struct {
			Tunnels []struct {
				URL string `json:"public_url"`
			} `json:"tunnels"`
		}
		err = jsonDec.Decode(&res)
		if err != nil {
			if verbose {
				log.Println("Error decoding JSON from tunnels requrest. Error: ", err)
			}
			continue
		}
		switch len(res.Tunnels) {
		case 0:
			if verbose {
				log.Println("Did not find a tunnel in the request", err)
			}
			continue
		case 1:
			// connection established! send it on its way!
			url := res.Tunnels[0].URL
			if verbose {
				log.Println("NGROK tunnel sucessfully established at ", url)
			}
			waitForConnectionChan <- url
			return
		default: //len > 1
			//there is more than one connection. I don't know how to handle it! Aborting...
			fmt.Fprintln(os.Stderr, "Error: more than one ngrok tunnel detected.")
			waitForConnectionChan <- ""
			return
		}
	}
}
