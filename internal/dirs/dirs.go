package dirs

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
)

//go:embed css/* assets/*
var static embed.FS

// staticServer serves front-end resources from [static].
var staticServer http.Handler

//go:embed html/*
var templ embed.FS

// t is the template used to render directory listings.
var t *template.Template = template.New(".")

func init() {
	t = t.Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format(time.RFC1123)
		},
		"formatSize": func(size int64) string {
			return datasize.ByteSize(size).HumanReadable()
		},
		"formatMode": func(fm fs.FileMode) (res string, err error) {
			res = fm.String()
			if len(res) == 0 {
				return "", errors.New("invalid file mode")
			}

			return res[1:], nil
		},
	})

	var err error
	if t, err = t.ParseFS(templ, "html/dir.gohtml"); err != nil {
		panic(err)
	}

	staticServer = http.StripPrefix("/", http.FileServer(http.FS(static)))
}

// Dirs is a proxy for http.FileServer that handles directory listings and file
// uploads.
type Dirs struct {
	http.Handler
	http.FileSystem
}

// ServeHTTP implements the [http.Handler] interface for *Dirs.
func (h *Dirs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := fs.Stat(static, strings.TrimPrefix(r.URL.Path, "/")); err == nil {
		// Serve static files.
		staticServer.ServeHTTP(w, r)

		return
	}

	if !h.respond(w, r) {
		// Handle the request with default Handler.
		h.Handler.ServeHTTP(w, r)
	}
}

// respond tries to handle the request and returns true if the request was
// handled.  It only handles the requests for directory listings and uploading
// the files.
func (h *Dirs) respond(w http.ResponseWriter, r *http.Request) (handled bool) {
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
