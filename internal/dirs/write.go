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

	"filesrv/internal/ferrors"
)

// ErrUnhandled is returned when the request is not handled by the upload
// handler.
const ErrUnhandled ferrors.Str = "unhandled request"

// handleUpload handles the upload of a multipart file from r.  It uses the
// URL's path as the directory for storing the file.
func (h *dirs) handleUpload(w http.ResponseWriter, r *http.Request, dst string) (err error) {
	if !r.URL.Query().Has("upload") {
		return fmt.Errorf("dirs: upload: %w", ErrUnhandled)
	}

	err = r.ParseMultipartForm(h.maxUploadSize)
	if err != nil {
		return fmt.Errorf("dirs: parsing multipart form: %w", err)
	}

	file, handler, err := r.FormFile("upload_file")
	if err != nil {
		return fmt.Errorf("dirs: retrieving file: %w", err)
	}
	defer file.Close()

	err = saveFile(file, handler, dst)
	if err != nil {
		return fmt.Errorf("dirs: %w", err)
	}

	log.Printf("Uploaded File: %q", handler.Filename)
	log.Printf("File Size:     %d", handler.Size)
	log.Printf("MIME Header:   %s", handler.Header)

	return nil
}

// saveFile saves the uploaded file to the given directory path dst.
func saveFile(file multipart.File, handler *multipart.FileHeader, dst string) (err error) {
	var tmpName string
	if ext := filepath.Ext(handler.Filename); ext != "" {
		tmpName = handler.Filename[:len(handler.Filename)-len(ext)] + "_*" + ext
	} else {
		tmpName = handler.Filename + "_*"
	}

	dstDir := filepath.Join(".", dst)
	f, err := os.CreateTemp(dstDir, tmpName)
	if err != nil {
		return err
	}
	defer closeAndRename(&err, f, filepath.Join(dstDir, handler.Filename))

	written, err := io.Copy(f, file)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	} else if written != handler.Size {
		return fmt.Errorf("wrote %d bytes, expected %d", written, handler.Size)
	}

	return nil
}

// closeAndRename renames the temporary file f to the final name if the caller
// succeeded.  Otherwise, it deletes the temporary file.  In both cases, it adds
// the own error to the caller's error.  Note that it closes the file even if
// the caller failed.  callerErr must not be nil (*callerErr could).
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
