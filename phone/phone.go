//Package phone abstracts away sending various message types to a phone.
package phone

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/mfresonke/ngrokker"
)

//Sender holds the necessary values for sending supported data to a phone.
//
//Must be initalized with the NewSender func.
type Sender struct {
	verbose bool
	port    int
	tunnel  ngrokker.Tunneler
}

//NewSender creates a new sender object with the specified options. Utilizes the
// ngrokker pkg, and in turn, ngrok, for its introspective tunneling purposes,
// and Twilio for its MMS sending purposes. If you'd like to override these
// services with something else (or provide alternate options to those services)
// see the NewCustomSender method.
//
//Users must accept the ngrok and Twilio ToS before sending anything.
//
//Port will be used to create a local webserver and introspective tunnel, if
// necessary. The port must not be currently in use by another process.
//
//Verbose prints diagnostic information to stderr.
func NewSender(
	acceptedNGROKTOS bool,
	port int,
	verbose bool,
) *Sender {
	tunnel := ngrokker.NewHTTPTunnel(acceptedNGROKTOS, verbose)
	return NewCustomSender(tunnel, port, verbose)
}

//NewCustomSender is similar to NewSender, except that it allows you to
// override the introspective tunneling service with your own.
//This is also useful for testing purposes.
func NewCustomSender(
	tunnel ngrokker.Tunneler,
	port int,
	verbose bool,
) *Sender {
	return &Sender{
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

	// open the introspective tunnel
	tunnel := ngrokker.NewHTTPTunnel(s.acceptedNGROKTOS, s.verbose)
	_, err := tunnel.Open(s.port)
	if err != nil {
		return err
	}

	return nil
}

var isValidPhotoExtRegex = regexp.MustCompile(".*(.jpg|.jpeg|.gif|.png|.bmp)")

func isValidPhotoExt(fileExtension string) bool {
	return isValidPhotoExtRegex.MatchString(fileExtension)
}
