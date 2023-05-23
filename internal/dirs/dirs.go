package dirs

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
)

//go:embed html/* css/* assets/*
var static embed.FS
var staticServer http.Handler

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
	if t, err = t.ParseFS(static, "html/dir.gohtml"); err != nil {
		panic(err)
	}

	staticServer = http.StripPrefix("/", http.FileServer(http.FS(static)))
}

type Dirs struct {
	http.Handler
	http.FileSystem
}

func (h *Dirs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := fs.Stat(static, strings.TrimPrefix(r.URL.Path, "/")); err == nil {
		staticServer.ServeHTTP(w, r)

		return
	}

	if !h.respond(w, r) {
		h.Handler.ServeHTTP(w, r)
	}
}

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
		h.handleDir(w, r)

		return true
	}

	return false
}
