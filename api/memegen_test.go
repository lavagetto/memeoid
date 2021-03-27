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

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const baseUrl string = "url"
const baseImgPath string = "../img/fixtures/"
const fontName string = "DejaVuSans"


type MemeGenTestSuite struct {
	suite.Suite
	TempDir string
	Sut *MemeHandler 
}

func (s *MemeGenTestSuite) SetupSuite() {
	tempdir, err := ioutil.TempDir("", "memeoid-api")
	if err != nil {
		panic(err)
	}
	os.Mkdir(path.Join(tempdir, baseUrl), os.FileMode(0755))
	s.TempDir = tempdir
}

func (s *MemeGenTestSuite) TeardownSuite() {
	os.RemoveAll(s.TempDir)
}

func (s *MemeGenTestSuite) SetupTest() {
	s.Sut = &MemeHandler{
		OutputPath: s.TempDir,
		ImgPath:    baseImgPath,
		FontName:   fontName,
		MemeURL:    baseUrl,
	}
}

func (s *MemeGenTestSuite) TeardownTest() {
	s.Sut = nil
}

func (s *MemeGenTestSuite) TestUID() {
	// Two requests with the same parameters create the same UID
	reader := strings.NewReader("")
	r := httptest.NewRequest(http.MethodGet, "http://localhost/w/api.php?first=a&last=b", reader)
	r1 := httptest.NewRequest(http.MethodGet, "http://localhost/w/api.php?last=b&first=a", reader)
	
	uid, err := s.Sut.UID(r)
	s.Nil(err, "could not calculate the UID: %v", err)
	
	uid1, err := s.Sut.UID(r1)
	s.Nil(err, "could not calculate the UID: %v", err)
	s.Equal(uid, uid1, "expected the UIDs to be equal for the same query parameters")
	
	// But this is case-sensitive.
	r1 = httptest.NewRequest(http.MethodGet, "http://localhost/w/api.php?last=b&First=a", reader)
	
	uid1, err = s.Sut.UID(r1)
	s.Nil(err, "could not calculate the UID: %v", err)
	s.NotEqual(uid, uid1, "expected the UIDs to be different for different capitalizations")
}

func (s *MemeGenTestSuite) TestListGifs() {
	originalPath := s.Sut.ImgPath
	var testCases = []struct {
		Path        string
		ContentType string
		Status      string
		Body        string
	}{
		{"", "application/json", "200 OK", `["badfile.gif","earth.gif"]`},
		{"/nonexistent", "", "404 Not Found", ""},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("Path: %s - ContentType: %s - Status: %s - Body: %s", tc.Path, tc.ContentType, tc.Status, tc.Body)
		s.Run(testName, func(){
			req := httptest.NewRequest(http.MethodGet, "http://localhost/gifs", strings.NewReader(""))
			req.Header.Set("Accept", "text/json")
			rec := httptest.NewRecorder()
			if tc.Path != "" {
				s.Sut.ImgPath = tc.Path
			} else {
				s.Sut.ImgPath = originalPath
			}
			s.Sut.ListGifs(rec, req)
			response := rec.Result()
			if tc.ContentType != "" {
				s.Equal([]string{tc.ContentType}, response.Header["Content-Type"])
			}
			if tc.Body != "" {
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				s.Nil(err)
				s.Equal(tc.Body, string(body))
			}
			s.Equal(tc.Status, response.Status)
		})
	}
}

func (s *MemeGenTestSuite) TestMemeGenerate() {
	var testCases = []struct {
		Uri           string
		StatusCode    int
		FileGenerated bool
	}{
		{"http://localhost/w/api.php", http.StatusBadRequest, false},
		{"http://localhost/w/api.php?from=lala", http.StatusNotFound, false},
		{"http://localhost/w/api.php?from=earth.gif", http.StatusBadRequest, false},
		{"http://localhost/w/api.php?from=earth.gif&top=test", http.StatusPermanentRedirect, true},
	}
	for _, tc := range testCases {
		testName := fmt.Sprintf("Uri: %s - StatusCode: %d - Genereate: %t", tc.Uri, tc.StatusCode, tc.FileGenerated)
		s.Run(testName, func() {
			req := httptest.NewRequest(http.MethodGet, tc.Uri, strings.NewReader(""))
			rec := httptest.NewRecorder()
	
			s.Sut.MemeFromRequest(rec, req)
			response := rec.Result()
	
			s.Equal(tc.StatusCode, response.StatusCode)
			
			if tc.FileGenerated {
				uid, err := s.Sut.UID(req)
				s.Nil(err)
				filePath := path.Join(s.Sut.OutputPath, fmt.Sprintf("%s.gif", uid))
				s.FileExists(filePath)
				s.Contains(response.Header["Location"][0], fmt.Sprintf("/url/%s", uid))
			}
		})
	}
}

func TestMemeGenTestSuite(t *testing.T) {
    suite.Run(t, new(MemeGenTestSuite))
}
