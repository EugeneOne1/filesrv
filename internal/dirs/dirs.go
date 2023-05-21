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
	"golang.org/x/exp/slices"
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

	if !h.respond(w, r.URL.Path) {
		h.Handler.ServeHTTP(w, r)
	}
}

func (h *Dirs) respond(w http.ResponseWriter, path string) (handled bool) {
	if !strings.HasSuffix(path, "/") {
		return false
	}

	f, err := h.Open(path)
	if err != nil {
		log.Printf("dirs: %s", err)

		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Printf("dirs: %s", err)

		return false
	} else if !fi.IsDir() {
		return false
	}

	entries, err := f.Readdir(-1)
	if err != nil {
		log.Printf("dirs: %s", err)

		return false
	}

	slices.SortFunc(entries, func(i, j fs.FileInfo) bool {
		if i.IsDir() {
			if j.IsDir() {
				return i.Name() < j.Name()
			}

			return true
		} else if j.IsDir() {
			return false
		}

		return i.Name() < j.Name()
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Lookup("dir.gohtml").Execute(w, struct {
		Path    string
		Entries []fs.FileInfo
	}{
		Path:    path,
		Entries: entries,
	})
	if err != nil {
		log.Printf("dirs: executing template: %s", err)
	}

	return true
}
