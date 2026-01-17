package toolkit

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	randStrBytes       = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
	randStrLen         = len(randStrBytes)
	defaultMaxFileSize = 1024 * 1024 * 1024

	rng = rand.NewPCG(
		uint64(time.Now().UnixNano()),
		uint64(time.Now().UnixNano()),
	)
	mu sync.Mutex
)

type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

func (t *Tools) RandomString(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)

	mu.Lock()
	defer mu.Unlock()

	var val uint64
	var bits uint = 0

	for i := 0; i < n; {
		if bits < 6 {
			val = rng.Uint64()
			bits = 64
		}

		idx := int(val & 0x3F)
		val >>= 6
		bits -= 6

		if idx == 63 {
			continue
		}

		b[i] = randStrBytes[idx]
		i++
	}

	return string(b)
}

type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	uploadedFiles, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}

	return uploadedFiles[0], nil
}

func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = defaultMaxFileSize
	}

	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("uploaded file is too big")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func([]*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile

				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buf := make([]byte, 512)
				_, err = infile.Read(buf)
				if err != nil {
					return nil, err
				}

				allowed := false
				fileType := http.DetectContentType(buf)

				if len(t.AllowedFileTypes) > 0 {
					for _, t := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, t) {
							allowed = true
							break
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("uploaded file type is not permitted")
				}

				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf(
						"%s%s",
						t.RandomString(25),
						filepath.Ext(hdr.Filename),
					)
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				}

				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}

				uploadedFile.FileSize = fileSize
				uploadedFile.OriginalFileName = hdr.Filename
				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}
