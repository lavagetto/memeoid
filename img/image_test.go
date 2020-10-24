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
	"image"
	"image/color"
	"testing"

	"github.com/flopp/go-findfont"
	"github.com/fogleman/gg"
)

func loadFont() string {
	// Tests will need dejavu sans to be installed on the system
	fontPath, err := findfont.Find("DejaVuSans.ttf")
	if err != nil {
		panic(err)
	}
	return fontPath
}

var textBoxFontSizeTests = []struct {
	width  int
	height int
	size   float64
	hasErr bool
}{
	{0, 100, 0.0, true},       // one of the textbox dimensions is too small
	{1000, 1000, 52.0, false}, // a large box keeps the original size
	{200, 200, 42.0, false},   // A smaller box reduces the size
	{200, 25, 34.0, false},    // A thinner box avoids word wrapping
	{20, 20, 0.0, true},       // a too small box can't contain the text
}

func TestTexboxFontSize(t *testing.T) {
	fontpath := loadFont()
	text := "testing memeoid"
	for _, testCase := range textBoxFontSizeTests {
		box := TextBox{
			Txt:      &text,
			Width:    testCase.width,
			Height:   testCase.height,
			Center:   image.Point{0, 0},
			FontPath: fontpath,
		}
		fontSize, err := box.CalculateFontSize(52.0, 8.0)
		if testCase.hasErr {
			if err == nil {
				t.Errorf("Expected the box %dx%d to generate an error, got a fontsize of %f", testCase.width, testCase.height, fontSize)
			}
		} else {
			if err != nil {
				t.Errorf("For test case %v got unexpected error %v", testCase, err)
			}
			if fontSize != testCase.size {
				t.Errorf("Expected the box %dx%d to have font size %d, found %d",
					testCase.width, testCase.height, int(testCase.size), int(fontSize))
			}
		}
	}
}

func TestDrawText(t *testing.T) {
	fontpath := loadFont()
	// Create a simple textbox
	box := TextBox{
		Width:            300,
		Height:           200,
		Center:           image.Point{150, 100},
		FontPath:         fontpath,
		LineSpacingRatio: 0.2,
	}
	err := box.SetText("X", 52.0, 8.0)
	if err != nil {
		t.Errorf("Could not create the textbox: %v", err)
	}

	ctx := gg.NewContext(box.Width, box.Height)
	ctx.SetRGB(1, 0, 0)
	ctx.Clear()
	ctx.SetRGB(0, 0, 0)

	ctx.LoadFontFace(fontpath, box.FontSize)
	err = box.DrawText(ctx)
	if err != nil {
		t.Errorf("Could not draw the text, %v", err)
	}
	rendered := ctx.Image()
	// Test a pixel at the border of the box is not printed to
	red := color.RGBA{255, 0, 0, 255}
	originColor := rendered.At(0, 0)
	if originColor != red {
		t.Errorf("Expected red at 0,0; found %v", originColor)
	}
	// The color at the center should be white, as we're printing text there
	centerColor := rendered.At(150, 100)
	white := color.RGBA{255, 255, 255, 255}
	if centerColor != white {
		t.Errorf("Expected the box center to be white, found %v", centerColor)
	}
}

// TODO: add test for Memegen and Template
