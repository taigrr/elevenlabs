package encode

import (
	"fmt"
	"io"
	"mime/multipart"
)

// Body produces a fresh request body for each attempt.
type Body struct {
	ContentType string
	// Open returns the body for the next attempt.
	Open func() (io.Reader, error)
	// Replayable reports whether Open may be called more than once.
	Replayable bool
}

// FormBody streams a multipart body of form (a `form`-tagged struct) through
// an io.Pipe. Uploads are replayable only when every file reader is an
// io.Seeker.
func FormBody(form any) *Body {
	files := CollectFiles(form)
	offsets := make([]int64, len(files))
	replayable := true
	for i, f := range files {
		seeker, ok := f.Reader.(io.Seeker)
		if !ok {
			replayable = false
			continue
		}
		off, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			replayable = false
			continue
		}
		offsets[i] = off
	}
	boundary := multipart.NewWriter(io.Discard).Boundary()
	attempts := 0
	open := func() (io.Reader, error) {
		if attempts > 0 {
			for i, f := range files {
				if _, err := f.Reader.(io.Seeker).Seek(offsets[i], io.SeekStart); err != nil {
					return nil, fmt.Errorf("elevenlabs: rewind upload: %w", err)
				}
			}
		}
		attempts++
		pr, pw := io.Pipe()
		go func() {
			mw := multipart.NewWriter(pw)
			_ = mw.SetBoundary(boundary)
			err := WriteForm(mw, form)
			if err == nil {
				err = mw.Close()
			}
			pw.CloseWithError(err)
		}()
		return pr, nil
	}
	return &Body{
		ContentType: "multipart/form-data; boundary=" + boundary,
		Open:        open,
		Replayable:  replayable,
	}
}
