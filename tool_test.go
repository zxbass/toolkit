package toolkit

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)

	if len(s) != 10 {
		t.Error("wrong random string length")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	rename        bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, rename: false, errorExpected: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, rename: true, errorExpected: false},
	{name: "filetype not allowed", allowedTypes: []string{"image/png"}, rename: false, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := new(sync.WaitGroup)
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/cat.jpg")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/cat.jpg")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error encoding image:", err)
			}

			err = jpeg.Encode(part, img, nil)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.rename)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s expected file to exist %s", e.name, err.Error())
			}

			os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Error("error expected none received")
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := new(sync.WaitGroup)
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/cat.jpg")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/cat.jpg")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error encoding image:", err)
			}

			err = jpeg.Encode(part, img, nil)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFile, err := testTools.UploadFile(request, "./testdata/uploads/", e.rename)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s expected file to exist %s", e.name, err.Error())
			}

			os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Error("error expected none received")
		}

		wg.Wait()
	}
}

func TestTools_CreateDirIfNotExists(t *testing.T) {
	var tools Tools

	if err := tools.CreateDirIfNotExists("./testdata/testdir"); err != nil {
		t.Error(err)
	}

	if err := tools.CreateDirIfNotExists("./testdata/testdir"); err != nil {
		t.Error(err)
	}

	if _, err := os.Stat("./testdata/testdir"); os.IsNotExist(err) {
		t.Error(err)
	}

	os.Remove("./testdata/testdir")
}
