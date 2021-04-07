package img

/*
Copyright © 2020 Giuseppe Lavagetto

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"image"
	"image/gif"
	"os"
	"strconv"
	"strings"

	"github.com/flopp/go-findfont"
)

// MemeTemplate represents all the basic
// information you need to generate a meme:
// the base image and the position, shape and
// font sizes in the text boxes.
type MemeTemplate struct {
	gifPath     string
	fontName    string
	Boxes       []TextBox
	border      float64
	minFontSize float64
	maxFontSize float64
	lineSpacing float64
}

// GetGif reads the gif from disk
func (tpl *MemeTemplate) GetGif() (*gif.GIF, error) {
	r, err := os.Open(tpl.gifPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return gif.DecodeAll(r)
}

// GetMeme fills a template with the text strings provided
func (tpl *MemeTemplate) GetMeme(text ...string) (*Meme, error) {
	numText := len(text)
	numBoxes := len(tpl.Boxes)
	// Copy the textboxes, we definitely don't want to deal with concurrency issues
	memeBoxes := make([]TextBox, numBoxes)
	if numText != numBoxes {
		return nil, fmt.Errorf("%d text pieces were given, but %d expected", numText, numBoxes)
	}
	for i, box := range tpl.Boxes {
		err := box.SetText(text[i], tpl.maxFontSize, tpl.minFontSize)
		if err != nil {
			return nil, err
		}
		memeBoxes[i] = box
	}
	g, err := tpl.GetGif()
	meme := Meme{Gif: g, TextBoxes: &memeBoxes, Border: tpl.border}
	return &meme, err
}

// SimpleTemplate generates the simplest possible template:
// - one box in the top 1/3rd of the image
// - one box in the bottom 1/3rd of the image
func SimpleTemplate(imgPath string, fontName string, maxFontSize float64, minFontSize float64) (*MemeTemplate, error) {
	fontPath, err := findfont.Find(fontName)
	if err != nil {
		return nil, err
	}
	tpl := MemeTemplate{
		gifPath:     imgPath,
		minFontSize: minFontSize,
		maxFontSize: maxFontSize,
		border:      0.01,
		lineSpacing: 0.3,
	}
	// We need the size of the image
	g, err := tpl.GetGif()
	if err != nil {
		return &tpl, err
	}
	// Now generate the textboxes
	size := g.Image[0].Bounds()
	imgWidth := float64(size.Dx())
	imgHeight := float64(size.Dy())
	width := imgWidth * (1.0 - 2.0*tpl.border)
	height := imgHeight * (1.0/3.0 - tpl.border)
	X := int(imgWidth * 0.5)
	Y := int(imgHeight*tpl.border + height*0.5)
	topBox := TextBox{
		Width:            int(width),
		Height:           int(height),
		Center:           image.Point{X, Y},
		FontPath:         fontPath,
		LineSpacingRatio: tpl.lineSpacing,
	}
	Y = int(imgHeight - imgHeight*tpl.border - height*0.5)
	bottomBox := TextBox{
		Width:            int(width),
		Height:           int(height),
		Center:           image.Point{X, Y},
		FontPath:         fontPath,
		LineSpacingRatio: tpl.lineSpacing,
	}
	tpl.Boxes = []TextBox{topBox, bottomBox}
	return &tpl, err
}

// MemeFromFile initiates a simple meme from a gif
func MemeFromFile(path string, top string, bottom string, fontName string) (*Meme, error) {
	tpl, err := SimpleTemplate(path, fontName, 52.0, 8.0)
	if err != nil {
		return nil, err
	}
	return tpl.GetMeme(top, bottom)
}

type BoxData struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Length int `json:"length"`
}

func NewBoxDataFromQuery(enc string) (BoxData, error) {
	var int_values [4]int
	for i, val := range strings.Split(enc, "|") {
		v, err := strconv.Atoi(val)
		if err != nil {
			return BoxData{}, err
		}
		int_values[i] = v
	}
	return BoxData{
		X:      int_values[0],
		Y:      int_values[1],
		Width:  int_values[2],
		Length: int_values[3],
	}, nil
}

type MemeRequest struct {
	From  string    `json:"from"`
	Boxes []BoxData `json:"boxes"`
	Texts []string  `json:"texts"`
}

// MemeFromRequest generates a meme using the template data provided.
func MemeFromRequest(memeReq *MemeRequest, fontName string) (*Meme, error) {
	fontPath, err := findfont.Find(fontName)
	if err != nil {
		return nil, err
	}
	tpl := MemeTemplate{
		gifPath:     memeReq.From,
		minFontSize: 8.0,
		maxFontSize: 52.0,
		border:      0.01,
		lineSpacing: 0.3,
	}
	for _, box := range memeReq.Boxes {
		txt := TextBox{
			Width:            box.Width,
			Height:           box.Length,
			Center:           image.Point{box.X, box.Y},
			FontPath:         fontPath,
			LineSpacingRatio: tpl.lineSpacing,
		}
		tpl.Boxes = append(tpl.Boxes, txt)
	}
	return tpl.GetMeme(memeReq.Texts...)
}
