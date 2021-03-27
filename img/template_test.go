package img

import (
	"fmt"
	"image"
	"image/gif"
	"testing"

	"github.com/flopp/go-findfont"
	"github.com/stretchr/testify/suite"
)

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

func (s *TemplateTestSuite) TestGetMeme() {
	box := TextBox{Width: 100, Height: 50, Center: image.Point{200, 200}, FontPath: s.fontPath}
	tpl := MemeTemplate{
		gifPath:     "fixtures/earth.gif",
		boxes:       []TextBox{box},
		border:      0.1,
		minFontSize: 8.0,
		maxFontSize: 52.0,
		lineSpacing: 0.2,
	}
	// Meme with no text provided => error
	m, err := tpl.GetMeme()
	s.Error(err, "no text provided should cause a failure")
	m, err = tpl.GetMeme("test")
	s.Nil(err, "error loading the meme: %v", err)
	s.Equal(*(*m.TextBoxes)[0].Txt, "test", "Not correctly assigned text to textbox")
	s.Equal((*m.TextBoxes)[0].FontSize, 52.0, "Not correctly set the font size")
	s.IsType(&gif.GIF{}, m.Gif, "A gif should be loaded")
	tpl.gifPath = "fixtures/badfile.gif"
	m, err = tpl.GetMeme("test")
	s.Error(err, "Should not generate a meme if the gif is corrupted")
}

func (s *TemplateTestSuite) TestGetGif() {
	var testGetGif = []struct {
		path string
		err  bool
	}{
		{"fixtures/earth.gif", false},
		{"fixtures/non-existent.gif", true},
		{"fixtures/badfile.gif", true},
	}
	for _, tc := range testGetGif {
		testName := fmt.Sprintf("path: %s - err: %t", tc.path, tc.err)
		s.Run(testName, func() {
			m := MemeTemplate{gifPath: tc.path}
			_, err := m.GetGif()
			if tc.err && err == nil {
				s.Errorf(nil, "test loading %s should have generated an error: %v", tc.path, err)
			}
			if !tc.err && err != nil {
				s.Errorf(err, "test loading %s should have not generated an error: %v", tc.path, err)
			}
		})
	}
}

func TestTemplateTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateTestSuite))
}
