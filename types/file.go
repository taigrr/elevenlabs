package types

import "io"

// File is an upload for multipart endpoints. Reader is streamed to the
// server without being buffered in memory. If Reader also implements
// io.Seeker the request can be retried; otherwise retries are disabled for
// that request.
type File struct {
	// Reader supplies the file contents.
	Reader io.Reader
	// Name is the filename reported to the server.
	Name string
	// ContentType is the part's MIME type; it defaults to application/octet-stream.
	ContentType string
}

// NewFile is shorthand for File{Reader: r, Name: name}.
func NewFile(r io.Reader, name string) File {
	return File{Reader: r, Name: name}
}
