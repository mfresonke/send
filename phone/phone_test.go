package phone

import (
	"os"
	"testing"
)

const (
	testingPort        = 7070
	testingVerbose     = false
	testingValidNumber = "+14071111111"
)

func TestErrFileDoesNotExist(t *testing.T) {
	sender := NewSender(testingPort, testingVerbose)
	err := sender.SendFile(testingValidNumber, "/this/does/not/exist")
	if err == nil {
		t.Fatal("No error returned")
	}
	if err != ErrFileDoesNotExist {
		t.Error("Putting an invalid file did not properly return ErrFileDoesNotExist")
	}
}

func TestErrFiletypeNotSupported(t *testing.T) {
	invalidFilePath := os.TempDir() + "/invalidFile.lol"
	_, err := os.Create(invalidFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(invalidFilePath)
	sender := NewSender(testingPort, testingVerbose)
	err = sender.SendFile(testingValidNumber, invalidFilePath)
	if err == nil {
		t.Fatal("No error returned")
	}
	if err != ErrFiletypeNotSupported {
		t.Error("Putting an invalid file did not properly return ErrFiletypeNotSupported")
	}
}
