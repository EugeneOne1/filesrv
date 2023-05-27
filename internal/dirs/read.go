package dirs

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"filesrv/internal/errors"

	"golang.org/x/exp/slices"
)

var ErrNotDir = errors.Error("not a directory")

// handleDir reads the directory and marshals the entries via the template.
func (h *Dirs) handleDir(w http.ResponseWriter, r *http.Request) (err error) {
	entries, err := h.read(r.URL.Path)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
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

	err = renderPage(w, r, entries)
	if err != nil {
		return fmt.Errorf("rendering page: %w", err)
	}

	return nil
}

type pathPart struct {
	Dir  string
	Path string
}

func renderPage(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo) (err error) {
	dirs := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	parts := make([]pathPart, 0, len(dirs))

	current := "/"
	for i := range dirs {
		parts = append(parts, pathPart{
			Dir:  dirs[i],
			Path: strings.Join(dirs[:i+1], "/") + "/",
		})
		current = parts[i].Dir
	}

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Lookup("dir.gohtml").Execute(w, struct {
		CurrentDir string
		PathParts  []pathPart
		Path       string
		Entries    []fs.FileInfo
	}{
		CurrentDir: current,
		PathParts:  parts,
		Path:       r.URL.Path,
		Entries:    entries,
	})
	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

// read reads the directory named by path and returns a list of directory
// entries.
func (h *Dirs) read(path string) (entries []fs.FileInfo, err error) {
	f, err := h.Open(path)
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
