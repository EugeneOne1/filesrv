package dirs

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
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

func renderPage(w http.ResponseWriter, r *http.Request, entries []fs.FileInfo) (err error) {
	p := strings.TrimSuffix(r.URL.Path, "/")
	parentDir, currentDir := path.Split(p)
	if currentDir == "" {
		parentDir, currentDir = "", parentDir
	} else {
		parentDir = strings.TrimSuffix(parentDir, "/")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Lookup("dir.gohtml").Execute(w, struct {
		ParentDir  string
		CurrentDir string
		Path       string
		Entries    []fs.FileInfo
	}{
		ParentDir:  parentDir,
		CurrentDir: currentDir,
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
