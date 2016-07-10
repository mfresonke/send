package phone

import (
	"os"
	"path/filepath"
	"regexp"
)

//Sender holds the necessary values for sending supported data to a phone.
//
//Must be initalized with the NewSender func.
type Sender struct {
	verbose bool
	port    int
}

//NewSender creates a sender object with the specified options.
//
//port must be a port that is not currently in use by another process.
//
//verbose prints diagnostic information to stderr.
func NewSender(port int, verbose bool) *Sender {
	return &Sender{
		port:    port,
		verbose: verbose,
	}
}

//SendFile sends a file to the specified phone number.
//
//Currently, it only supports photos, but support for additional files
// is planned.
func (sender Sender) SendFile(phoneNumber, filePath string) error {
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
	return nil
}

var isValidPhotoExtRegex = regexp.MustCompile(".*(.jpg|.jpeg|.gif|.png|.bmp)")

func isValidPhotoExt(fileExtension string) bool {
	return isValidPhotoExtRegex.MatchString(fileExtension)
}
