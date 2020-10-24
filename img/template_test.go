package img

import (
	"image"
	"image/gif"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testGetGif = []struct {
	path string
	err  bool
}{
	{"fixtures/earth.gif", false},
	{"fixtures/non-existent.gif", true},
	{"fixtures/badfile.gif", true},
}

func TestGetGif(t *testing.T) {
	for _, test := range testGetGif {
		m := MemeTemplate{gifPath: test.path}
		_, err := m.GetGif()
		if test.err && err == nil {
			t.Errorf("test loading %s should have generated an error: %v", test.path, err)
		}
		if !test.err && err != nil {
			t.Errorf("test loading %s should have not generated an error: %v", test.path, err)
		}
	}
}

func TestGetMeme(t *testing.T) {
	box := TextBox{Width: 100, Height: 50, Center: image.Point{200, 200}, FontPath: loadFont()}
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
	assert.Error(t, err, "no text provided should cause a failure")
	m, err = tpl.GetMeme("test")
	assert.Nil(t, err, "error loading the meme: %v", err)
	assert.Equal(t, *(*m.TextBoxes)[0].Txt, "test", "Not correctly assigned text to textbox")
	assert.Equal(t, (*m.TextBoxes)[0].FontSize, 52.0, "Not correctly set the font size")
	assert.IsType(t, &gif.GIF{}, m.Gif, "A gif should be loaded")
	tpl.gifPath = "fixtures/badfile.gif"
	m, err = tpl.GetMeme("test")
	assert.Error(t, err, "Should not generate a meme if the gif is corrupted")
}
