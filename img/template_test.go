package img

import (
	"fmt"
	"image"
	"image/gif"
	"testing"

	"github.com/flopp/go-findfont"
	"github.com/stretchr/testify/suite"
)

// TODO: this struct is the same as image_test.go:ImageTestSuite
// Refactoring might be taken into account in order to reduce code duplication
type TemplateTestSuite struct {
	suite.Suite
	fontPath string
}

func (s *TemplateTestSuite) SetupSuite() {
	fontPath, err := findfont.Find(defaultFont)
	if err != nil {
		panic(err)
	}
	s.fontPath = fontPath
}

func (s TemplateTestSuite) createTemplate() MemeTemplate {
	box := TextBox {
		Width:    100,
		Height:   50,
		Center:   image.Point{200, 200},
		FontPath: s.fontPath,
	}
	return MemeTemplate {
		gifPath:     "fixtures/earth.gif",
		boxes:       []TextBox{box},
		border:      0.1,
		minFontSize: 8.0,
		maxFontSize: 52.0,
		lineSpacing: 0.2,
	}
}

func (s *TemplateTestSuite) TestGetMemeWithNoProvidedName() {
	sut := s.createTemplate()

	_, err := sut.GetMeme()

	s.Error(err, "no text provided should cause a failure")
}

func (s *TemplateTestSuite) TestGetMeme() {
	sut := s.createTemplate()

	m, err := sut.GetMeme("test")

	s.Nil(err, "error loading the meme: %v", err)
	s.Equal(*(*m.TextBoxes)[0].Txt, "test", "Not correctly assigned text to textbox")
	s.Equal((*m.TextBoxes)[0].FontSize, 52.0, "Not correctly set the font size")
	s.IsType(&gif.GIF{}, m.Gif, "A gif should be loaded")
}

func (s *TemplateTestSuite) TestCorruptedSourceFile() {
	sut := s.createTemplate()
	sut.gifPath = "fixtures/badfile.gif"
	
	_, err := sut.GetMeme("test")

	s.Error(err, "Should not generate a meme if the gif is corrupted")
}

func (s *TemplateTestSuite) TestGetGif() {
	var testGetGif = []struct {
		path     string
		isError  bool
	}{
		{"fixtures/earth.gif", false},
		{"fixtures/non-existent.gif", true},
		{"fixtures/badfile.gif", true},
	}
	for _, tc := range testGetGif {
		testName := fmt.Sprintf("path: %s - err: %t", tc.path, tc.isError)
		s.Run(testName, func() {
			sut := MemeTemplate{ gifPath: tc.path }
			
			_, err := sut.GetGif()
			
			if tc.isError && err == nil {
				s.Errorf(nil, "test loading %s should have generated an error: %v", tc.path, err)
			}
			if !tc.isError && err != nil {
				s.Errorf(err, "test loading %s should have not generated an error: %v", tc.path, err)
			}
		})
	}
}

func TestTemplateTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateTestSuite))
}
