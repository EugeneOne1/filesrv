package dirs

import (
	"log"
	"net/http"
	"path"
	"strings"
)

// indexPage is the suffix of the index file's name.
const indexPage = "/index.html"

// ServeHTTP implements the [http.Handler] interface for *dirs.
func (h *dirs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/") {
		r.URL.Path = "/" + r.URL.Path
	}

	// Redirect .../index.html to .../.
	if strings.HasSuffix(r.URL.Path, indexPage) {
		localRedirect(w, r, "./")

		return
	}

	h.serveFile(w, r, path.Clean(r.URL.Path))
}

// serveFile serves the file under name to w.
func (h *dirs) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	f, err := h.fsys.Open(name)
	if err != nil {
		staticFile, staticErr := h.theme.Open(name)
		if staticErr != nil {
			h.theme.RenderError(w, r, err)

			return
		}

		f = staticFile
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Printf("closing served file %q: %v", name, err)
		}
	}()

	d, err := f.Stat()
	if err != nil {
		h.theme.RenderError(w, r, err)

		return
	}

	// Redirect to canonical path: "/" at end of directory p, [r.URL.Path]
	// always begins with "/".
	p := r.URL.Path
	if d.IsDir() {
		if !strings.HasSuffix(p, "/") {
			localRedirect(w, r, path.Base(p)+"/")

			return
		}

		// Use contents of index.html for directory, if present.
		ff, err := h.fsys.Open(strings.TrimSuffix(name, "/") + indexPage)
		if err == nil {
			defer ff.Close()

			dd, err := ff.Stat()
			if err == nil {
				d = dd
				f = ff
			}
		}
	} else if strings.HasSuffix(p, "/") {
		localRedirect(w, r, "../"+path.Base(p))

		return
	}

	if d.IsDir() {
		// Still a directory, no index.html.
		h.handleDir(w, r, d)
	} else {
		http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	}
}

// localRedirect gives an [http.StatusMovedPermanently] response.  It does not
// convert relative paths to absolute paths like [http.Redirect] does, for
// example when [http.StripPrefix] is used.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}
