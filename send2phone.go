package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
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
	if opts.Verbose {
		fmt.Println("Searching for ngrok in path...")
	}
	ngrokLoc, err := exec.LookPath("ngrok")
	check(err)
	if opts.Verbose {
		fmt.Println("ngrok found at ", ngrokLoc)
	}
	//check for config
	_, err = loadConfig()
	check(err)

	// ====== Application Logic ======
	// run ngrok to be sure there are no issues
	runNGROK(opts.Port)

	// check the extension of the given file to make sure it is compatible with twilio
	fileExt := filepath.Ext(opts.Args.Input)
	if ok := isValidPhotoExt(fileExt); !ok {
		ErrPrintln("Error, image filetype is not supported.")
		os.Exit(1)
	}

	time.Sleep(10 * time.Second)

}

func runNGROK(port int) {
	cmd := exec.Command("ngrok", "http", strconv.Itoa(port))
	cmd.Start()
	go func() {
		err := cmd.Wait()
		check(err)
	}()
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
