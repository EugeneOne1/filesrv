package themes

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"filesrv/internal/dirs"

	"github.com/c2h5oh/datasize"
)

//go:embed css/* assets/* html/*
var static embed.FS

type defaultTheme struct {
	templ  *template.Template
	static fs.FS
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

func (t *defaultTheme) Render(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo) {
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
	err := t.templ.Lookup("dir.gohtml").Execute(w, templData)
	if err != nil {
		log.Printf("%s: executing template: %v", t, err)
	}
}

// RenderError implements the [dirs.Theme] interface for *defaultTheme.
func (t *defaultTheme) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	templData := struct {
		Title      string
		Message    string
		Favicon    string
		StatusCode int
	}{}

	switch {
	case errors.Is(err, fs.ErrNotExist):
		templData.Message = "Requested resource isn't found."
		templData.Favicon = "🌚"
		templData.StatusCode = http.StatusNotFound
	case errors.Is(err, fs.ErrPermission):
		templData.Message = "You do not have permission to access the requested resource."
		templData.Favicon = "🔒"
		templData.StatusCode = http.StatusForbidden
	default:
		templData.Message = fmt.Sprintf("Something went wrong: %v.", err)
		templData.Favicon = "❌"
		templData.StatusCode = http.StatusInternalServerError
	}
	templData.Title = http.StatusText(templData.StatusCode)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.templ.Lookup("err.gohtml").Execute(w, templData)
	if err != nil {
		log.Printf("%s: executing template: %v", t, err)
	}
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

// DefaultEmbedded returns a new theme based on the embedded assets.
func DefaultEmbedded() (theme dirs.Theme) {
	t, err := template.New(".").Funcs(funcMap).ParseFS(static, "html/dir.gohtml", "html/err.gohtml")
	if err != nil {
		// This should never happen since the whole content is embedded.
		panic(err)
	}

	th := DefaultDynamic(static).(*defaultDynamic).defaultTheme
	th.templ = t

	return &th
}

// ServeHTTP implements the [http.Handler] interface for *defaultTheme.
func (t *defaultTheme) Open(name string) (f http.File, err error) {
	return http.FS(t.static).Open(name)
}

// String implements the [fmt.Stringer] interface for *defaultTheme.
func (t *defaultTheme) String() string {
	return fmt.Sprintf("Default[fs=%T]", t.static)
}

// defaultDynamic is a a wrapper around defaultTheme that parses the template on
// each request and serves static files from provided file system.
type defaultDynamic struct {
	defaultTheme
}

// DefaultDynamic returns a new theme based on fsys.  It parses the template
// from the fsys on each request.
func DefaultDynamic(fsys fs.FS) (theme dirs.Theme) {
	return &defaultDynamic{
		defaultTheme: defaultTheme{
			static: fsys,
		},
	}
}

// Render implements the [dirs.Theme] interface for *defaultDynamic.
func (d *defaultDynamic) Render(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo) {
	(&defaultTheme{
		templ: template.Must(template.New(r.Host).
			Funcs(funcMap).
			ParseFS(d.static, "html/dir.gohtml"),
		),
		static: d.static,
	}).Render(w, r, entries)
}

// RenderError implements the [dirs.Theme] interface for *defaultDynamic.
func (d *defaultDynamic) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	(&defaultTheme{
		templ: template.Must(template.New(r.Host).
			Funcs(funcMap).
			ParseFS(d.static, "html/err.gohtml"),
		),
		static: d.static,
	}).RenderError(w, r, err)
}

// String implements the [fmt.Stringer] interface for *defaultDynamic.
func (d defaultDynamic) String() string {
	return fmt.Sprintf("Default[fs=%T]", d.static)
}
