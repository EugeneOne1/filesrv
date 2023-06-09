package dirs

import (
	"fmt"
	"io/fs"
	"net/http"
)

// Theme is the interface for the directory listing appearance.
type Theme interface {
	// Render renders the HTML page using entries and info from the request.
	Render(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo)

	// RenderNotFound renders the [http.StatusNotFound] page.  It should be
	// ready to handle [ErrUnhandled].
	RenderError(w http.ResponseWriter, r *http.Request, err error)

	// http.FileSystem is embedded here to allow theme serve its static content.
	// Any request that accesses the static content has already been failed to
	// be handled by the served file system.
	http.FileSystem

	// fmt.Stringer is embedded here to allow theme being named.
	fmt.Stringer
}

// dirs is an [http.Handler] that handles directory listings and file uploads.
type dirs struct {
	fsys          http.FileSystem
	theme         Theme
	maxUploadSize int64
}

// HTTPFSConfig is the configuration for creating file listings handler.
type HTTPFSConfig struct {
	// FS is the http.FileSystem used to serve actual files.
	FS http.FileSystem

	// Theme is the theme used to render the directory listings.
	Theme Theme

	// MaxUploadSize is the maximum size of a file that can be uploaded in
	// bytes.
	MaxUploadSize int64
}

// NewHTTPFSDirs creates a new [http.Handler] that handles directory listings
// and file uploads.
func NewHTTPFSDirs(conf *HTTPFSConfig) (d http.Handler, err error) {
	return &dirs{
		fsys:          conf.FS,
		theme:         conf.Theme,
		maxUploadSize: conf.MaxUploadSize,
	}, nil
}
