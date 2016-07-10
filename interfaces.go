package main

// FileSender represents a type that can send a photo to some place.
// The interface is both input and destination-agnostic
type FileSender interface {
	SendFile(destination, filePath string) error
}
