package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mfresonke/ngrokker"
)

const debug = true

type phoneNumber string

type options struct {
	// Example of positional arguments
	Args struct {
		Input string
	} `positional-args:"yes"`
	// SaveNumber string `short:"s" long:"save" description:"save a phone number"`
	Verbose bool `short:"v" long:"verbose" description:"Show more information about what is happening"`
	Port    int  `default:"7070" short:"p" long:"port" description:"Port on which to run the temporary webserver for twilio."`
}

func main() {
	// ====== Setup ======
	// parse the cmdline opts
	var opts options
	_, err := flags.Parse(&opts)
	check(err)
	// always be verbose if debug is enabled.
	if debug {
		opts.Verbose = true
	}
	// start the w
	// open the introspective tunnel
	tunnel := ngrokker.NewHTTPTunnel(true, opts.Verbose)
	url, err := tunnel.Open(opts.Port)
	check(err)
	fmt.Println(url)

	//check for config
	_, err = loadConfig()
	check(err)

	// ====== Application Logic ======

}

func check(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

type phoneConfig struct {
	Numbers map[string]phoneNumber
	Default string
}

func (pc phoneConfig) defaultNumber() phoneNumber {
	return pc.Numbers[pc.Default]
}

func loadConfig() (phoneConfig, error) {
	//fake it for now
	return phoneConfig{
		Default: "maxs-phone",
		Numbers: map[string]phoneNumber{
			"maxs-phone": "+14075758643",
		},
	}, nil
}

// ErrPrintln prints a string to standard error.
func ErrPrintln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}
