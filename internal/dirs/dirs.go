package dirs

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

// Theme is the interface for the directory listing themes.
type Theme interface {
	// Render renders the HTML page using entries and info from the request.
	Render(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo) (err error)

	// [http.Handler] is embedded here to allow a theme to serve its static
	// files or some dynamic content.
	http.Handler

	IsContentRequest(r *http.Request) (ok bool)
}

// dirs is a proxy for http.FileServer that handles directory listings and file
// uploads.
type dirs struct {
	fsys        http.FileSystem
	httpDefault http.Handler

	theme Theme
}

type HTTPFSConfig struct {
	// FS is the http.FileSystem used to serve actual files.
	FS http.FileSystem

	// Theme is the theme used to render the directory listings.
	Theme Theme
}

func NewHTTPFSDirs(conf *HTTPFSConfig) (d http.Handler, err error) {
	return &dirs{
		fsys:        conf.FS,
		httpDefault: http.FileServer(conf.FS),
		theme:       conf.Theme,
	}, nil
}

// ServeHTTP implements the [http.Handler] interface for *Dirs.
//
// TODO(e.burkov):  Reimplement this method with at least some basic security
// considerations in mind.
func (h *dirs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.theme.IsContentRequest(r) {
		h.theme.ServeHTTP(w, r)

		return
	}

	if !h.respond(w, r) {
		// Handle the request with default Handler.
		h.httpDefault.ServeHTTP(w, r)
	}
}

// respond tries to handle the request and returns true if the request was
// handled.  It only handles the requests for directory listings and uploading
// the files.
func (h *dirs) respond(w http.ResponseWriter, r *http.Request) (handled bool) {
	if !strings.HasSuffix(r.URL.Path, "/") {
		return false
	}

	switch r.Method {
	case http.MethodPost:
		if r.URL.Query().Has("upload") {
			err := h.handleUpload(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return true
			}

			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)

			return true
		}
	case http.MethodGet:
		err := h.handleDir(w, r)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return false
			}

			log.Printf("error handling directory: %v", err)
			// http.Error(w, err.Error(), http.StatusInternalServerError)

			return true
		}

		return true
	}

	return false
}
