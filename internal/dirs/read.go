package dirs

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

// handleDir reads the directory and marshals the entries via the template.
func (h *dirs) handleDir(w http.ResponseWriter, r *http.Request, f http.File, d fs.FileInfo) {
	mtime := d.ModTime()

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		ims := r.Header.Get("If-Modified-Since")
		if ims != "" && !isZeroTime(mtime) {
			t, err := http.ParseTime(ims)
			if err == nil && !mtime.After(t) && mtime.Truncate(time.Second).Compare(t) <= 0 {
				writeUnmodified(w)

				return
			}
		}
	case http.MethodPost:
		err := h.handleUpload(w, r, d.Name())
		if err != nil {
			h.theme.RenderError(w, r, err)
		} else {
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		}

		return
	}

	if !isZeroTime(mtime) {
		w.Header().Set("Last-Modified", mtime.UTC().Format(http.TimeFormat))
	}

	entries, err := h.readdir(r, f)
	if err != nil {
		h.theme.RenderError(w, r, fmt.Errorf("reading directory: %w", err))

		return
	}

	h.theme.Render(w, r, entries)
}

type doubleDot struct {
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (c *doubleDot) Name() string       { return ".." }
func (c *doubleDot) Size() int64        { return c.size }
func (c *doubleDot) Mode() fs.FileMode  { return c.mode }
func (c *doubleDot) ModTime() time.Time { return c.modTime }
func (c *doubleDot) IsDir() bool        { return true }
func (c *doubleDot) Sys() any           { return nil }

func (h *dirs) readdir(r *http.Request, f http.File) (entries []fs.FileInfo, err error) {
	entries, err = f.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	if parentPath := path.Dir(strings.TrimRight(r.URL.Path, "/")); parentPath != "." {
		var parent http.File
		parent, err = h.fsys.Open(parentPath)
		if err != nil {
			return nil, fmt.Errorf("opening parent directory: %w", err)
		}
		defer func() {
			err = parent.Close()
			if err != nil {
				err = fmt.Errorf("closing parent directory: %w", err)
			}
		}()

		var parentInfo fs.FileInfo
		parentInfo, err = parent.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat parent directory: %w", err)
		}

		entries = append([]fs.FileInfo{&doubleDot{
			mode:    parentInfo.Mode(),
			modTime: parentInfo.ModTime(),
			size:    parentInfo.Size(),
		}}, entries...)
	}

	return entries, nil
}

// writeUnmodified writes a [http.StatusNotModified] response.
func writeUnmodified(w http.ResponseWriter) {
	// RFC 7232 section 4.1:
	// a sender SHOULD NOT generate representation metadata other than the
	// above listed fields unless said metadata exists for the purpose of
	// guiding cache updates (e.g., Last-Modified might be useful if the
	// response does not have an ETag field).
	h := w.Header()
	delete(h, "Content-Type")
	delete(h, "Content-Length")
	delete(h, "Content-Encoding")
	if h.Get("Etag") != "" {
		delete(h, "Last-Modified")
	}
	w.WriteHeader(http.StatusNotModified)
}

// unixEpochTime is the time.Time corresponding to the Unix epoch (00:00:00 UTC
// on 1 January 1970).
var unixEpochTime = time.Unix(0, 0)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}
