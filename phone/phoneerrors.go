package phone

import "errors"

//ErrFiletypeNotSupported is returned when the given input file is a valid
// file, but its extension is not supported by the sender service.
//
//Examples of invalid types include ".exe", ".docx", and ".pptx"
var ErrFiletypeNotSupported = errors.New("The given filetype is not supported")

//ErrFileDoesNotExist is returned when the given input file cannot be found
var ErrFileDoesNotExist = errors.New("The given file path does not exist.")

//ErrFileIsDirectory is returned when the given filepath is a directory
var ErrFileIsDirectory = errors.New("The given file path is a directory")
