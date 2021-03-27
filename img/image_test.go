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
	"image/color"
	"testing"

	"github.com/flopp/go-findfont"
	"github.com/fogleman/gg"
	"github.com/stretchr/testify/suite"
)

const defaultFont string = "DejaVuSans.ttf"

type ImageTestSuite struct {
	suite.Suite
	fontPath string
}

func (s *ImageTestSuite) SetupSuite() {
	fontPath, err := findfont.Find(defaultFont)
	if err != nil {
		panic(err)
	}
	s.fontPath = fontPath
}

func (s *ImageTestSuite) TestDrawText() {
	// Create a simple textbox
	box := TextBox{
		Width:            300,
		Height:           200,
		Center:           image.Point{150, 100},
		FontPath:         s.fontPath,
		LineSpacingRatio: 0.2,
	}
	err := box.SetText("X", 52.0, 8.0)
	if err != nil {
		s.Errorf(err, "Could not create the textbox: %v", err)
	}

	ctx := gg.NewContext(box.Width, box.Height)
	ctx.SetRGB(1, 0, 0)
	ctx.Clear()
	ctx.SetRGB(0, 0, 0)

	ctx.LoadFontFace(s.fontPath, box.FontSize)
	err = box.DrawText(ctx)
	if err != nil {
		s.Errorf(err, "Could not draw the text, %v", err)
	}
	rendered := ctx.Image()
	// Test a pixel at the border of the box is not printed to
	red := color.RGBA{255, 0, 0, 255}
	originColor := rendered.At(0, 0)
	if originColor != red {
		s.Errorf(nil, "Expected red at 0,0; found %v", originColor)
	}
	// The color at the center should be white, as we're printing text there
	centerColor := rendered.At(150, 100)
	white := color.RGBA{255, 255, 255, 255}
	if centerColor != white {
		s.Errorf(nil, "Expected the box center to be white, found %v", centerColor)
	}
}

func (s *ImageTestSuite) TestTexboxFontSize() {
	text := "testing memeoid"
	var testCases = []struct {
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
	for _, tc := range testCases {
		testName := fmt.Sprintf("width: %d - height: %d - size: %f - hasErr: %t", tc.width, tc.height, tc.size, tc.hasErr)
		s.Run(testName, func(){
			box := TextBox{
				Txt:      &text,
				Width:    tc.width,
				Height:   tc.height,
				Center:   image.Point{0, 0},
				FontPath: s.fontPath,
			}
			fontSize, err := box.CalculateFontSize(52.0, 8.0)
			if tc.hasErr {
				if err == nil {
					s.Errorf(err, "Expected the box %dx%d to generate an error, got a fontsize of %f", tc.width, tc.height, fontSize)
				}
			} else {
				if err != nil {
					s.Errorf(err, "For test case %v got unexpected error %v", tc, err)
				}
				if fontSize != tc.size {
					s.Errorf(nil, "Expected the box %dx%d to have font size %d, found %d",
						tc.width, tc.height, int(tc.size), int(fontSize))
				}
			}
		})
	}
}

// TODO: add test for Memegen and Template

func TestImageTestSuite(t *testing.T) {
	suite.Run(t, new(ImageTestSuite))
}


