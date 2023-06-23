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

	// redirect to canonical path: / at end of directory p
	// r.URL.Path always begins with /
	p := r.URL.Path
	if d.IsDir() {
		if !strings.HasSuffix(p, "/") {
			localRedirect(w, r, path.Base(p)+"/")

			return
		}
	} else {
		if strings.HasSuffix(p, "/") {
			localRedirect(w, r, "../"+path.Base(p))

			return
		}
	}

	if d.IsDir() {
		url := r.URL.Path
		// redirect if the directory name doesn't end in a slash
		if url == "" || url[len(url)-1] != '/' {
			localRedirect(w, r, path.Base(url)+"/")

			return
		}

		// use contents of index.html for directory, if present
		index := strings.TrimSuffix(name, "/") + indexPage
		ff, err := h.fsys.Open(index)
		if err == nil {
			defer ff.Close()

			dd, err := ff.Stat()
			if err == nil {
				d = dd
				f = ff
			}
		}
	}

	// Still a directory, no index.html.
	if d.IsDir() {
		h.handleDir(w, r, d)

		return
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
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
