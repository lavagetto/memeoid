package img

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/fogleman/gg"
)

// Meme is a structure describing a meme
type Meme struct {
	// The gif.GIF for the image
	Gif *gif.GIF
	// Text to add at the top
	Top *TextBox
	// Text to add at the bottom
	Bottom *TextBox
	// Border (fraction of image size)
	Border float64
}

//TextBox represents a text box to add to the image.
type TextBox struct {
	// Text to write in the textbox
	Txt *string
	// Width of the textbox, in pixels
	Width int
	// Height of the textbox
	Height int
	// Position of the textbox in the image
	Center image.Point
	// Path to the font
	FontPath string
	// Line spacing (fraction of the fontsize)
	LineSpacingRatio float64
	// the actual font size.
	FontSize float64
}

func NewTextBox(text string, w int, h int, center image.Point, fontPath string, maxFontSize float64, minFontSize float64, lineSpacing float64) (*TextBox, error) {
	t := TextBox{
		Txt:              &text,
		Width:            w,
		Height:           h,
		Center:           center,
		FontPath:         fontPath,
		LineSpacingRatio: lineSpacing,
	}
	fs, err := t.CalculateFontSize(maxFontSize, minFontSize)
	if err != nil {
		return nil, err
	}
	t.FontSize = fs
	return &t, nil
}

// CalculateFontSize calculates the maximum font size that can fit
// the text in the textbox.
func (t *TextBox) CalculateFontSize(maxFontSize float64, minFontSize float64) (float64, error) {
	if t.Height <= 0 || t.Width <= 0 {
		return 0.0, fmt.Errorf("image size is too small")
	}
	ctx := gg.NewContext(t.Width, t.Height)
	for fs := maxFontSize; fs >= minFontSize; fs -= 2.0 {
		err := ctx.LoadFontFace(t.FontPath, fs)
		if err != nil {
			return 0.0, fmt.Errorf("font at %s could not be loaded at font size %d", t.FontPath, int(fs))
		}
		// Wrap text to fit maxWidth
		lines := ctx.WordWrap(*t.Txt, float64(t.Width))
		// This is a peculiarity of the gg API. Oh, well :P
		width, height := ctx.MeasureMultilineString(strings.Join(lines, "\n"), math.Ceil(ctx.FontHeight()*t.LineSpacingRatio))
		if int(width) <= t.Width && int(height) <= t.Height {
			return fs, nil
		}
	}
	return 0.0, fmt.Errorf("text can't fit in the image")
}

// DrawText draws the text into a gg context
func (t *TextBox) DrawText(ctx *gg.Context) error {
	if *t.Txt == "" {
		return fmt.Errorf("trying to draw an empty string.")
	}
	// Stroke size needs to be 40% of the line spacing.
	lineSpacing := math.Ceil(ctx.FontHeight() * t.LineSpacingRatio)
	strokeSize := int(lineSpacing * 0.4)
	// Write the text with a white fill and black stroke.
	// Inspired by the meme.go example in gg
	ctx.SetHexColor("#000")
	for dy := -strokeSize; dy <= strokeSize; dy++ {
		for dx := -strokeSize; dx <= strokeSize; dx++ {
			// Round corners (x^2 + y2 < r^2)
			if dx*dx+dy*dy >= strokeSize*strokeSize {
				continue
			}
			ctx.DrawStringAnchored(*t.Txt, float64(t.Center.X+dx), float64(t.Center.Y+dy), 0.5, 0.5)
		}
	}
	ctx.SetHexColor("#FFF")
	ctx.DrawStringAnchored(*t.Txt, float64(t.Center.X), float64(t.Center.Y), 0.5, 0.5)
	return nil
}

// MemeFromFile initiates a meme from a gif
func MemeFromFile(path string, top string, bottom string, fontPath string) (*Meme, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	g, err := gif.DecodeAll(r)
	if err != nil {
		return nil, err
	}
	// TODO: get these from config
	border := 0.01
	maxFontSize := 52.0
	minFontSize := 8.0
	lineSpacing := 0.2
	// Now generate the textboxes
	size := g.Image[0].Bounds()
	imgWidth := float64(size.Dx())
	imgHeight := float64(size.Dy())
	width := imgWidth * (1.0 - 2.0*border)
	height := imgHeight * (1.0/3.0 - border)
	X := int(imgWidth * 0.5)
	Y := int(imgHeight*border + height*0.5)
	c := image.Point{X, Y}
	topBox, err := NewTextBox(top, int(width), int(height), c, fontPath, maxFontSize, minFontSize, lineSpacing)
	if err != nil {
		return nil, err
	}
	c.Y = int(imgHeight - imgHeight*border - height*0.5)
	bottomBox, err := NewTextBox(bottom, int(width), int(height), c, fontPath, maxFontSize, minFontSize, lineSpacing)
	if err != nil {
		return nil, err
	}
	return &Meme{g, topBox, bottomBox, border}, nil
}

// GifMetaData returns metadata on the gif
func (m *Meme) GifMetaData() {
	for i, img := range m.Gif.Image {
		fmt.Println("##")
		fmt.Printf("Frame: %d\n", i)
		fmt.Printf("Delay: %dms\n", m.Gif.Delay[i]*10)
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		fmt.Printf("Size: %dx%d\n", width, height)
		fmt.Printf("Disposal: %b\n", m.Gif.Disposal[i])
	}
}

// NormalizeImage modifies the image it so that it has
// all frames of the same size, with transparency.
func (m *Meme) NormalizeImage() {
	g := m.Gif
	size := g.Image[0].Bounds()
	lastFullSizeIdx := 0
	for i, img := range g.Image {
		if img.Bounds() != size {
			// create an image of size "size", then draw the current image centered.
			newImage := image.NewPaletted(size, img.Palette)
			// First draw the last fullsize image here
			lastFullSize := g.Image[lastFullSizeIdx]
			draw.Draw(newImage, size, lastFullSize, image.Point{0, 0}, draw.Src)
			draw.Draw(newImage, img.Bounds(), img, img.Bounds().Min, draw.Src)
			// now swap the image
			g.Image[i] = newImage
		} else {
			lastFullSizeIdx = i
		}
	}
}

// Generate modifies the image adding the meme text
func (m *Meme) Generate() error {
	// We normalize the image as I'm not sure how drawing only on a fraction of the full gif would work.
	// This might be revisited later for more space-efficient generated gifs
	m.NormalizeImage()
	// process every frame in a goroutine
	var wg sync.WaitGroup
	var mux sync.Mutex
	for i, img := range m.Gif.Image {
		wg.Add(1)
		go m.drawTextAt(i, img, &wg, &mux)
	}
	// Wait for all frames to be rendered
	wg.Wait()
	return nil
}

func (m *Meme) drawTextAt(i int, img *image.Paletted, wg *sync.WaitGroup, mux *sync.Mutex) {
	size := img.Bounds()
	// Build a GG context, and load the font face
	ctx := gg.NewContext(size.Dx(), size.Dy())
	ctx.DrawImage(img, 0, 0)
	if *m.Top.Txt != "" {
		ctx.LoadFontFace(m.Top.FontPath, m.Top.FontSize)
		m.Top.DrawText(ctx)
	}
	if *m.Bottom.Txt != "" {
		ctx.LoadFontFace(m.Bottom.FontPath, m.Bottom.FontSize)
		m.Bottom.DrawText(ctx)
	}
	// Now we need to get a palettedimage back
	paletted := image.NewPaletted(img.Bounds(), img.Palette)
	draw.Draw(paletted, paletted.Rect, ctx.Image(), img.Bounds().Min, draw.Src)
	mux.Lock()
	m.Gif.Image[i] = paletted
	mux.Unlock()
	wg.Done()

}
