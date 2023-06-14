package dirs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/c2h5oh/datasize"
)

// maxFileSize is the maximum size of a file that can be uploaded.  It's 4GB for
// now.
const maxFileSize = int64(4 * datasize.GB)

// handleUpload handles the upload of a multipart file from r.  It uses the
// URL's path as the directory for storing the file.
func (h *dirs) handleUpload(r *http.Request) (err error) {
	r.ParseMultipartForm(maxFileSize)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("upload_file")
	if err != nil {
		return fmt.Errorf("dirs: retrieving file: %w", err)
	}
	defer file.Close()

	err = saveFile(file, handler, r.URL.Path)
	if err != nil {
		return fmt.Errorf("dirs: %w", err)
	}

	log.Printf("Uploaded File: %q", handler.Filename)
	log.Printf("File Size:     %d", handler.Size)
	log.Printf("MIME Header:   %s", handler.Header)

	return nil
}

// filenamePattern returns a pattern for [os.CreateTemp].  It essentially places
// the wildcard to the end of the name, but before the extension (if any).
func filenamePattern(name string) (pattern string) {
	ext := filepath.Ext(name)
	if ext == "" {
		return name + "_*"
	}

	return name[:len(name)-len(filepath.Ext(name))] + "_*" + ext
}

// saveFile saves the uploaded file to the given directory path.
func saveFile(file multipart.File, handler *multipart.FileHeader, to string) (err error) {
	dir := filepath.Join(".", to)
	fname := filepath.Join(dir, handler.Filename)

	f, err := os.CreateTemp(dir, filenamePattern(handler.Filename))
	if err != nil {
		return err
	}
	defer closeAndRename(&err, f, fname)

	written, err := io.Copy(f, file)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	} else if written != handler.Size {
		return fmt.Errorf("wrote %d bytes, expected %d", written, handler.Size)
	}

	return nil
}

// closeAndRename renames the temporary file f to the final name if the caller
// succeeded.  Otherwise, it deletes the temporary file.  In both cases, it
// adds the own error to the caller's error.  Note that it closes the file even
// if the caller failed.  callerErr must not be nil (*callerErr could).
func closeAndRename(callerErr *error, f *os.File, finalName string) {
	// It's required on Windows to close the file before renaming it.
	err := f.Close()
	if err != nil {
		err = errors.Join(*callerErr, err)
	}

	var action string
	if err != nil {
		err = errors.Join(os.Remove(f.Name()), err)
		action = "removing temporary file"
	} else if finalName != "" {
		switch _, err = os.Lstat(finalName); {
		case err == nil:
			// File exists, leave the name as is.
		case !errors.Is(err, os.ErrNotExist):
			// Some other error, report it.
			action = "checking file existence"
		default:
			// File doesn't exist, rename the temporary file.
			err = os.Rename(f.Name(), finalName)
			action = "renaming temporary file"
		}
	} else {
		return
	}

	if err != nil {
		*callerErr = errors.Join(*callerErr, fmt.Errorf("%s: %w", action, err))
	}
}
