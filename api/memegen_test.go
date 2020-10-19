package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setUp() *MemeHandler {
	tempdir, err := ioutil.TempDir("", "memeoid-api")
	if err != nil {
		panic(err)
	}
	m := MemeHandler{
		OutputPath: tempdir,
		ImgPath:    "../img/fixtures/",
		FontName:   "DejaVuSans",
		MemeURL:    "url",
	}
	os.Mkdir(path.Join(tempdir, m.MemeURL), os.FileMode(755))
	return &m
}

func tearDown(m *MemeHandler) {
	os.RemoveAll(m.OutputPath)
}

func TestUID(t *testing.T) {
	m := setUp()
	defer tearDown(m)
	// Two requests with the same parameters create the same UID
	reader := strings.NewReader("")
	r := httptest.NewRequest(http.MethodGet, "http://localhost/w/api.php?first=a&last=b", reader)
	r1 := httptest.NewRequest(http.MethodGet, "http://localhost/w/api.php?last=b&first=a", reader)
	uid, err := m.UID(r)
	assert.Nil(t, err, "could not calculate the UID: %v", err)
	uid1, err := m.UID(r1)
	assert.Nil(t, err, "could not calculate the UID: %v", err)
	assert.Equal(t, uid, uid1, "expected the UIDs to be equal for the same query parameters")
	// But this is case-sensitive.
	r1 = httptest.NewRequest(http.MethodGet, "http://localhost/w/api.php?last=b&First=a", reader)
	uid1, err = m.UID(r1)
	assert.Nil(t, err, "could not calculate the UID: %v", err)
	assert.NotEqual(t, uid, uid1, "expected the UIDs to be different for different capitalizations")
}

var testListGifs = []struct {
	Path        string
	ContentType string
	Status      string
	Body        string
}{
	{"", "application/json", "200 OK", `["badfile.gif","earth.gif"]`},
	{"/nonexistent", "", "404 Not Found", ""},
}

func TestListGifs(t *testing.T) {
	h := setUp()
	defer tearDown(h)
	reader := strings.NewReader("")
	originalPath := h.ImgPath
	for _, test := range testListGifs {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/gifs", reader)
		rec := httptest.NewRecorder()
		if test.Path != "" {
			h.ImgPath = test.Path
		} else {
			h.ImgPath = originalPath
		}
		h.ListGifs(rec, req)
		response := rec.Result()
		if test.ContentType != "" {
			assert.Equal(t, []string{test.ContentType}, response.Header["Content-Type"])
		}
		if test.Body != "" {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			assert.Nil(t, err)
			assert.Equal(t, test.Body, string(body))
		}
		assert.Equal(t, test.Status, response.Status)
	}
}

var testMemeGenerate = []struct {
	Uri           string
	StatusCode    int
	FileGenerated bool
}{
	{"http://localhost/w/api.php", http.StatusBadRequest, false},
	{"http://localhost/w/api.php?from=lala", http.StatusNotFound, false},
	{"http://localhost/w/api.php?from=earth.gif", http.StatusBadRequest, false},
	{"http://localhost/w/api.php?from=earth.gif&top=test", http.StatusPermanentRedirect, true},
}

func TestMemeGenerate(t *testing.T) {
	h := setUp()
	defer tearDown(h)
	for _, test := range testMemeGenerate {
		reader := strings.NewReader("")
		req := httptest.NewRequest(http.MethodGet, test.Uri, reader)
		rec := httptest.NewRecorder()
		h.MemeFromRequest(rec, req)
		response := rec.Result()
		assert.Equal(t, test.StatusCode, response.StatusCode)
		if test.FileGenerated {
			uid, err := h.UID(req)
			assert.Nil(t, err)
			filePath := path.Join(h.OutputPath, "url", fmt.Sprintf("%s.gif", uid))
			assert.FileExists(t, filePath)
			assert.Contains(t, response.Header["Location"][0], uid)
		}
	}
}
