package themes

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"filesrv/internal/dirs"

	"github.com/c2h5oh/datasize"
)

//go:embed css/* assets/*
var static embed.FS

//go:embed html/*
var templ embed.FS

type defaultTheme struct {
	templ         *template.Template
	static        fs.FS
	staticHandler http.Handler
}

var _ dirs.Theme = (*defaultTheme)(nil)

// pathPart represents a part of a path to directory.
type pathPart struct {
	// Dir is the directory name with no slashes.  The root directory is
	// represented by an empty string.
	Dir string
	// Path is the full path containing slashes on both ends.  The root
	// directory is represented by a single slash.
	Path string
}

// pathParts returns the base directory name and a list of path parts.  The
// parts are in reversed order so that the root directory is the last element.
func pathParts(p string) (current string, parts []pathPart) {
	dirs := strings.Split(strings.TrimSuffix(p, "/"), "/")
	parts = make([]pathPart, 0, len(dirs))
	parts = append(parts, pathPart{
		Dir:  "",
		Path: "/",
	})
	current = "/"

	for i := range dirs[1:] {
		parts = append(parts, pathPart{
			Dir:  dirs[i+1],
			Path: strings.Join(dirs[:i+2], "/") + "/",
		})
		current = parts[i+1].Dir
	}

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return current, parts
}

func (t *defaultTheme) Render(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo) (err error) {
	sortBy(r.URL.Query().Get(paramSort), entries)

	templData := struct {
		CurrentDir string
		PathParts  []pathPart
		Path       string
		Params     url.Values
		Entries    []fs.FileInfo
	}{
		Path:    r.URL.Path,
		Params:  r.URL.Query(),
		Entries: entries,
	}
	templData.CurrentDir, templData.PathParts = pathParts(r.URL.Path)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.templ.Lookup("dir.gohtml").Execute(w, templData)
	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

var funcMap = template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format(time.RFC1123)
	},
	"formatSize": func(size int64) string {
		return datasize.ByteSize(size).HumanReadable()
	},
	"formatMode": func(fm fs.FileMode) (res string, err error) {
		buf := [12]byte{}
		bi := 0
		for i, c := range "rwxrwxrwx" {
			if fm&(1<<uint(9-1-i)) != 0 {
				buf[bi] = byte(c)
			} else {
				buf[bi] = '-'
			}
			if bi++; bi%3 == 0 && bi < len(buf) {
				buf[bi] = '\n'
				bi++
			}
		}

		return string(buf[:]), nil
	},
}

func DefaultEmbedded() (theme dirs.Theme) {
	t, err := template.New(".").Funcs(funcMap).ParseFS(templ, "html/dir.gohtml")
	if err != nil {
		// This should never happen since the whole content is embedded.
		panic(err)
	}

	return &defaultTheme{
		templ:         t,
		static:        static,
		staticHandler: http.StripPrefix("/", http.FileServer(http.FS(static))),
	}
}

func (t *defaultTheme) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.staticHandler.ServeHTTP(w, r)
}

func (t *defaultTheme) IsContentRequest(r *http.Request) (ok bool) {
	_, err := fs.Stat(t.static, strings.TrimPrefix(r.URL.Path, "/"))

	return err == nil
}

type defaultDynamic struct {
	defaultTheme
}

func DefaultDynamic(path string) (theme dirs.Theme) {
	ents, err := fs.Glob(os.DirFS(path), "*")
	if err != nil {
		panic(err)
	}

	log.Printf("entries: %v", ents)

	return defaultDynamic{
		defaultTheme: defaultTheme{
			static:        os.DirFS(path),
			staticHandler: http.StripPrefix("/", http.FileServer(http.FS(os.DirFS(path)))),
		},
	}
}

func (d defaultDynamic) Render(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo) (err error) {
	return (&defaultTheme{
		templ:         template.Must(template.New(r.Host).Funcs(funcMap).ParseFS(d.static, "html/dir.gohtml")),
		static:        d.static,
		staticHandler: d.staticHandler,
	}).Render(w, r, entries)
}

func (d defaultDynamic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.defaultTheme.ServeHTTP(w, r)
}

func (d defaultDynamic) IsContentRequest(r *http.Request) (ok bool) {
	return d.defaultTheme.IsContentRequest(r)
}
