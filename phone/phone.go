//Package phone abstracts away sending various message types to a phone.
package phone

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/mfresonke/ngrokker"
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
	// check that the given file exists and is not a directory
	if fileInfo, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileDoesNotExist
	} else if fileInfo.IsDir() {
		return ErrFileIsDirectory
	}

	// check the extension of the given file to make sure it is compatible with twilio
	fileExt := filepath.Ext(filePath)
	if ok := isValidPhotoExt(fileExt); !ok {
		return ErrFiletypeNotSupported
	}

	// start the go webserver to serve the image
	webserverErrChan := make(chan error, 1)
	go serveImage(webserverErrChan, s.port, filePath)

	//open the introspective tunnel
	_, err := s.tunnel.Open(s.port)
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

var isValidPhotoExtRegex = regexp.MustCompile(".*(.jpg|.jpeg|.gif|.png|.bmp)")

func isValidPhotoExt(fileExtension string) bool {
	return isValidPhotoExtRegex.MatchString(fileExtension)
}

func serveImage(errorChan chan error, port int, filePath string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	})
	bindStr := ":" + strconv.Itoa(port)
	errorChan <- http.ListenAndServe(bindStr, nil)
}
