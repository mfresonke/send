package main

import (
	"log"

	"github.com/mfresonke/send2phone/phone"

	"github.com/jessevdk/go-flags"
)

const debug = true

type phoneNumber string

type options struct {
	// Example of positional arguments
	Args struct {
		PhoneNumber string
		InputFile   string
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

	//check for config
	_, err = loadConfig()
	check(err)

	// ====== Application Logic ======
	twilCfg := phone.TwilioConfig{
	// some config here
	}
	sender := phone.NewSender(twilCfg, true, opts.Port, opts.Verbose)
	err = sender.SendFile(opts.Args.PhoneNumber, opts.Args.InputFile)
	check(err)
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
