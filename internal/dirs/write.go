package dirs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/c2h5oh/datasize"
)

const maxFileSize = int64(datasize.GB)

func (h *Dirs) handleUpload(r *http.Request) (err error) {
	r.ParseMultipartForm(maxFileSize)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("upload_file")
	if err != nil {
		return fmt.Errorf("dirs: retrieving file: %w", err)
	}
	defer file.Close()

	dir := filepath.Join(".", r.URL.Path)
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		return fmt.Errorf("dirs: creating temporary file: %w", err)
	}
	defer func(tmpName string) {
		var oerr error
		var action string
		if err != nil {
			oerr = os.Remove(tmpName)
			action = "removing"
		} else {
			oerr = os.Rename(tmpName, filepath.Join(dir, handler.Filename))
			action = "renaming"
		}
		if oerr != nil {
			log.Printf("dirs: %s temporary file: %s", action, oerr)
		}
	}(f.Name())
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return fmt.Errorf("dirs: writing file: %w", err)
	}

	log.Printf("Uploaded File: %q", handler.Filename)
	log.Printf("File Size:     %d", handler.Size)
	log.Printf("MIME Header:   %s", handler.Header)

	return nil
}
