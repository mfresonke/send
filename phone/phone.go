//Package phone abstracts away sending various message types to a phone.
package phone

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	twilio "github.com/carlosdp/twiliogo"
	"github.com/davecgh/go-spew/spew"
	"github.com/mfresonke/ngrokker"
)

const (
	// callback path for twilio requests
	twilioCallbackPath = "/callback"
	// the prefix used before hosting files. For more info see the "file" type.
	filePrefixPath = "/file"
)

//Sender holds the necessary values for sending supported data to a phone.
//
//Must be initalized with the NewSender func.
type Sender struct {
	verbose bool
	port    int
	tunnel  ngrokker.Tunneler
	config  TwilioConfig
}

//NewSender creates a new sender object with the specified options. Utilizes the
// ngrokker pkg, and in turn, ngrok, for its introspective tunneling purposes,
// and Twilio for its MMS sending purposes. If you'd like to override the tunneling
// service with something else, use NewSenderTunnel.
//
//A valid Twilio Configuration is needed to be able to properly send a message.
//
//Users must accept the ngrok ToS before sending anything.
//
//Port will be used to create a local webserver and introspective tunnel, if
// necessary. The port must not be currently in use by another process.
//
//Verbose prints diagnostic information to stderr.
func NewSender(
	config TwilioConfig,
	acceptedNGROKTOS bool,
	port int,
	verbose bool,
) *Sender {
	tunnel := ngrokker.NewHTTPTunnel(acceptedNGROKTOS, verbose)
	return NewSenderTunnel(tunnel, config, port, verbose)
}

//NewSenderTunnel is similar to NewSender, except that it allows you to
// override the introspective tunneling service with your own.
func NewSenderTunnel(
	tunnel ngrokker.Tunneler,
	config TwilioConfig,
	port int,
	verbose bool,
) *Sender {
	return &Sender{
		config:  config,
		tunnel:  tunnel,
		port:    port,
		verbose: verbose,
	}
}

//SendFile sends a file to the specified phone number.
//
//Currently, it only supports photos, but support for additional files
// is planned.
func (s Sender) SendFile(phoneNumber, filePath string) error {
	// validate the input file.
	file, err := newSendableFile(filePath)
	if err != nil {
		return err
	}

	// start the go webserver to serve the image
	webserverErrChan := make(chan error, 1)
	go serveFile(webserverErrChan, s.port, file, s.verbose)

	//open the introspective tunnel
	endpoints, err := s.tunnel.Open(s.port)
	if err != nil {
		return err
	}

	// find the first secure endpoint and use it as the base URL
	var secureEndpointURL string
	for _, ep := range endpoints {
		if ep.Secure {
			secureEndpointURL = ep.URL
			break
		}
	}
	// make sure we recieved a URL
	if secureEndpointURL == "" {
		return errors.New("no secure endpoint found")
	}

	//send the image!
	err = sendMMS(s.config, file, secureEndpointURL, phoneNumber)
	if err != nil {
		return err
	}

	// at some point check the channels for errors
	select {
	case _ = <-webserverErrChan:
		//do something useful
	}

	return nil
}

//sendMMS actually performs the Twilio MMS Request.
func sendMMS(cfg TwilioConfig, file sendableFile, baseURL string, phoneNumber string) error {
	client := twilio.NewClient(cfg.SID, cfg.AuthToken)
	params := []twilio.Optional{
		twilio.MediaUrl(file.publicURL(baseURL)),
		twilio.StatusCallback(baseURL + twilioCallbackPath),
	}
	msg, err := twilio.NewMessage(client, cfg.SenderNum, phoneNumber, params...)
	if err != nil {
		return err
	}
	spew.Dump(msg)
	return nil
}

//sendableFile represents a valid, sendable input file based on its path.
type sendableFile string

func newSendableFile(filePath string) (sendableFile, error) {
	// check that the given file exists and is not a directory
	if fileInfo, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", ErrFileDoesNotExist
	} else if fileInfo.IsDir() {
		return "", ErrFileIsDirectory
	}

	// check the extension of the given file to make sure it is compatible with twilio
	fileExt := filepath.Ext(filePath)
	if ok := isValidPhotoExt(fileExt); !ok {
		return "", ErrFiletypeNotSupported
	}
	// if all looks good, return a valid file object!
	return sendableFile(filePath), nil
}

func (f sendableFile) path() string {
	return string(f)
}

func (f sendableFile) name() string {
	_, fileName := path.Split(string(f))
	return fileName
}

func (f sendableFile) publicURL(baseURL string) string {
	fileName := f.name()
	urlFileName := url.QueryEscape(fileName)
	return baseURL + filePrefixPath + "/" + urlFileName
}

var isValidPhotoExtRegex = regexp.MustCompile(".*(.jpg|.jpeg|.gif|.png|.bmp)")

func isValidPhotoExt(fileExtension string) bool {
	return isValidPhotoExtRegex.MatchString(fileExtension)
}

func serveFile(errorChan chan error, port int, file sendableFile, verbose bool) {
	http.HandleFunc(twilioCallbackPath, func(w http.ResponseWriter, r *http.Request) {
		// implement twilio callback parsing here.
	})
	http.HandleFunc(filePrefixPath, func(w http.ResponseWriter, r *http.Request) {
		// converting from the the URL to the filename
		fileName := r.URL.Path[len(filePrefixPath)+1:]
		if fileName != file.name() {
			if verbose {
				log.Println(
					"unable to serve file, as fileName and file.name() differ. fileName:",
					fileName,
					"file.name():",
					file.name(),
				)
			}
			return
		}
		http.ServeFile(w, r, file.path())
	})
	bindStr := ":" + strconv.Itoa(port)
	errorChan <- http.ListenAndServe(bindStr, nil)
}
