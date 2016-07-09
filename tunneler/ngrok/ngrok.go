package ngrok

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mfresonke/send2phone/tunneler"
)

const (
	// how long to wait for ngrok before giving up
	connectionTimeout = 20 * time.Second
	// how long to wait before beginning to poll for a connection.
	initalConnectionWait = 5 * time.Second
	//urlPollHTTPTimeout   = 500 * time.Millisecond
	// how long to wait between poll attempts (given the previous one did not succeed)
	urlPollDuration         = 1 * time.Second
	sigtermPollDuration     = 200 * time.Millisecond
	sigtermRetriesUntilKill = 20
)

const tunnelsURL = "http://127.0.0.1:4040/api/tunnels"

type tunnel struct {
	*exec.Cmd
	verbose bool
}

// NewTunnel creates a new ngrok tunnel, ready to open!
func NewTunnel(verbose bool) tunneler.Interface {
	return &tunnel{
		verbose: verbose,
	}
}

func (tun *tunnel) Open(port int) ([]tunneler.Endpoint, error) {
	if tun.verbose {
		fmt.Println("Searching for ngrok in path...")
	}
	ngrokLoc, err := exec.LookPath("ngrok")
	if err != nil {
		return nil, err
	}
	if tun.verbose {
		fmt.Println("ngrok found at ", ngrokLoc)
	}

	// create and start ngrok!
	tun.Cmd = exec.Command("ngrok", "http", strconv.Itoa(port))
	err = tun.Start()
	if err != nil {
		return nil, err
	}

	// channel will send an error if the process exits unexpectedly.
	errorChan := make(chan error, 1)
	go func(errorChan chan error) {
		errorChan <- tun.Wait()
	}(errorChan)

	// channel will recieve the string of the connection URL.
	waitForConnectionChan := make(chan connectionInfo, 1)
	go connectionWaiter(waitForConnectionChan, tun.verbose)

	// and finally, make a channel that will time out if all else fails.
	timeoutChan := time.After(connectionTimeout)

	// wait for something to happen...
	var endpoints []tunneler.Endpoint
	select {
	case info := <-waitForConnectionChan:
		if info.err != nil {
			return nil, info.err
		}
		endpoints = info.endpoints
	case err := <-errorChan:
		return nil, err
	case <-timeoutChan:
		return nil, errors.New("NGROK startup timed out")
	}
	spew.Dump(endpoints)
	return endpoints, nil
}

func (tun *tunnel) Close() error {
	if tun.ProcessState.Exited() {
		return nil
	}
	if tun.verbose {
		log.Println("Sending SIGTERM to ngrok...")
	}
	tun.Process.Signal(syscall.SIGTERM)
	for i := 0; i != sigtermRetriesUntilKill; i++ {
		if tun.verbose {
			if tun.verbose {
				log.Println("Waiting for ngrok process to shutdown...", i+1)
			}
		}
		time.Sleep(sigtermPollDuration)
		if tun.ProcessState.Exited() {
			if tun.verbose {
				log.Println("NGROK shutdown sucessful.")
			}
			return nil
		}
	}
	if tun.verbose {
		log.Println("NGROK shutdown unsuccessful. Killing process.")
	}
	return tun.Process.Kill()
}

type connectionInfo struct {
	endpoints []tunneler.Endpoint
	err       error
}

// connectionWaiter pings the ngrok api until it discovers a connection. Once it
// does, it sends the resulting tunnel url on the channel.
// As the channel and no return imply, this func is meant to be run asyncronously.
func connectionWaiter(waitForConnectionChan chan connectionInfo, verbose bool) {
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
		jsonDec := json.NewDecoder(resp.Body)
		var res struct {
			Tunnels []struct {
				URL       string `json:"public_url"`
				Protocall string `json:"proto"`
			} `json:"tunnels"`
		}
		err = jsonDec.Decode(&res)
		// close the response body regardless if there was an error, since we are just
		//  going to "continue" the loop anyway.
		resp.Body.Close()
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
		case 2:
			// connection established! Create the endpoints!
			info := connectionInfo{}
			for _, tun := range res.Tunnels {
				isSecure := (tun.Protocall == "https")
				ep := tunneler.Endpoint{
					URL:    tun.URL,
					Secure: isSecure,
				}
				info.endpoints = append(info.endpoints, ep)
				if verbose {
					log.Println("NGROK tunnel sucessfully established at ", tun.URL)
				}
			}
			waitForConnectionChan <- info
			return
		default: //len > 1
			//there is more than one connection (2 tunnels == 1 connection).
			// I don't know how to handle it! Aborting...
			waitForConnectionChan <- connectionInfo{
				err: errors.New("Error: more than one ngrok tunnel detected."),
			}
			return
		}
	}
}
