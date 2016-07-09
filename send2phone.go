package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jessevdk/go-flags"
	"github.com/mfresonke/send2phone/tunneler/ngrok"
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
	var opts options
	_, err := flags.Parse(&opts)
	check(err)
	// always be verbose if debug is enabled.
	if debug {
		opts.Verbose = true
	}
	// check for dependencies
	tunnel := ngrok.NewTunnel(opts.Verbose)
	url, err := tunnel.Open(opts.Port)
	check(err)
	fmt.Println(url)

	//check for config
	_, err = loadConfig()
	check(err)

	// ====== Application Logic ======

	// check the extension of the given file to make sure it is compatible with twilio
	fileExt := filepath.Ext(opts.Args.Input)
	if ok := isValidPhotoExt(fileExt); !ok {
		ErrPrintln("Error, image filetype is not supported.")
		os.Exit(1)
	}

}

var isValidPhotoExtRegex = regexp.MustCompile(".*(.jpg|.jpeg|.gif|.png|.bmp)")

func isValidPhotoExt(fileExtension string) bool {
	return isValidPhotoExtRegex.MatchString(fileExtension)
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

type config struct {
	PhoneNumbers map[string]phoneNumber
	Default      string
}

func loadConfig() (config, error) {
	//fake it for now
	return config{
		Default: "maxs-phone",
		PhoneNumbers: map[string]phoneNumber{
			"maxs-phone": "+14075758643",
		},
	}, nil
}

// ErrPrintln prints a string to standard error.
func ErrPrintln(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}
