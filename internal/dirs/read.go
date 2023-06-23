package dirs

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

// read reads the directory named by path and returns a list of directory
// entries.
func (h *dirs) read(path string) (entries []fs.FileInfo, err error) {
	f, err := h.fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fs.ErrNotExist
	}

	entries, err = f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// handleDir reads the directory and marshals the entries via the template.
func (h *dirs) handleDir(w http.ResponseWriter, r *http.Request, d fs.FileInfo) {
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

	entries, err := h.read(r.URL.Path)
	if err != nil {
		h.theme.RenderError(w, r, fmt.Errorf("reading directory: %w", err))

		return
	}

	h.theme.Render(w, r, entries)
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
