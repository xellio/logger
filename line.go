package logger

//
// Line struct holds information about a single line in a file.
//
type Line struct {
	File    *File
	Content string
}

//
// NewLine creates and returns a new Line struct for the given content and  file.
//
func NewLine(content []byte, file *File) (*Line, error) {
	line := &Line{
		File:    file,
		Content: string(content),
	}
	return line, nil
}
