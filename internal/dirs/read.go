package dirs

import (
	"errors"
	"io/fs"
	"log"
	"net/http"

	"golang.org/x/exp/slices"
)

var ErrNotDir = errors.New("not a directory")

func (h *Dirs) handleDir(w http.ResponseWriter, r *http.Request) {
	entries, err := h.read(r.URL.Path)
	if err != nil {
		log.Printf("dirs: %s", err)

		return
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
		Path:    r.URL.Path,
		Entries: entries,
	})
	if err != nil {
		log.Printf("dirs: executing template: %s", err)
	}
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
		return nil, ErrNotDir
	}

	entries, err = f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
