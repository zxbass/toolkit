package toolkit

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
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

var slugTests = []struct {
	name       string
	s          string
	expected   string
	shoudlFail bool
}{
	{name: "normal string", s: "a string", expected: "a-string", shoudlFail: false},
	{name: "debil string", s: "a@%$%)string--$%($)", expected: "a-string", shoudlFail: false},
	{name: "empty string", s: "", shoudlFail: true},
	{name: "debiliest string", s: "&#^$%", shoudlFail: true},
}

func TestTools_Slugify(t *testing.T) {
	var tools Tools

	for _, test := range slugTests {
		slug, err := tools.Slugify(test.s)

		if test.shoudlFail && err == nil {
			t.Errorf("test should have failed for %s but it didn't. Got [%s]", test.s, slug)
		}

		if err != nil && !test.shoudlFail {
			t.Errorf("test should have passed for %s but it didn't", test.s)
		}

		if slug != test.expected {
			t.Errorf("slug %s expected, %s got", test.expected, slug)
		}
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	var tools Tools

	tools.DownloadStaticFile(rr, req, "./testdata", "cat.jpg", "image.jpg")

	res := rr.Result()
	defer res.Body.Close()

	if res.Header.Get("Content-Length") != "88614" {
		t.Errorf("wrong content length %s", res.Header.Get("Content-Length"))
	}

	if res.Header.Get("Content-Disposition") != "attachment; filename=\"image.jpg\"" {
		t.Errorf("wrong content disposition [%s]", res.Header.Get("Content-Disposition"))
	}

	_, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

}
