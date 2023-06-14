package dirs

import (
	"fmt"
	"io/fs"
	"net/http"
)

// read reads the directory named by path and returns a list of directory
// entries.
func (h *dirs) read(path string) (entries []fs.FileInfo, err error) {
	f, err := h.fsys.Open(path)
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

// handleDir reads the directory and marshals the entries via the template.
func (h *dirs) handleDir(w http.ResponseWriter, r *http.Request) (err error) {
	entries, err := h.read(r.URL.Path)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	err = h.theme.Render(w, r, entries)
	if err != nil {
		return fmt.Errorf("rendering page: %w", err)
	}

	return nil
}
