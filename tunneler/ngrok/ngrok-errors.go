package ngrok

import (
	"errors"
	"strings"
)

// these are a set of strings to match against the ngrok output text.
const (
	processOutputTooManyConnections = "is limited to 1 simultaneous ngrok client session."
)

// ErrTooManyConnections is returned when ngrok reports that it cannot start due
//  to another simultaneous connection. This generally means the ngrok process
//  is already running.
var ErrTooManyConnections = errors.New("NGROK cannot be started because there are too many simultaneous connections")

func newOutputError(outputBytes []byte) error {
	if len(outputBytes) == 0 {
		panic("newOutputError called with input of length 0")
	}
	output := string(outputBytes)
	var err error
	switch {
	case strings.Contains(output, processOutputTooManyConnections):
		err = ErrTooManyConnections
	default:
		err = errors.New("NGROK Outputted Text Unexpectedly. Text: " + output)
	}
	return err
}
