// Generates icon.ico with green "CLI" text for systray.
// Run: go run ./scripts/genicon
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/J-Siu/go-png2ico/v2/p2i"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	size     = 32
	text     = "CLI"
	greenHex = 0x00AA00
	fontSize = 18
)

func main() {
	// Parse font and create face
	parsed, err := opentype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		panic(err)
	}
	defer face.Close()

	// Create 32x32 RGBA with transparent background
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Green color
	green := color.RGBA{
		R: (greenHex >> 16) & 0xFF,
		G: (greenHex >> 8) & 0xFF,
		B: greenHex & 0xFF,
		A: 255,
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(green),
		Face: face,
	}
	advance := d.MeasureString(text)
	metrics := face.Metrics()
	width := advance.Round()
	height := (metrics.Ascent + metrics.Descent).Round()
	ascent := metrics.Ascent.Round()

	// Center horizontally: x = (size - width) / 2
	// Center vertically: baseline at size/2 + ascent - height/2
	x := fixed.I((size - width) / 2)
	y := fixed.I(size/2 + ascent - height/2)
	d.Dot = fixed.Point26_6{X: x, Y: y}
	d.DrawString(text)

	// Write PNG to temp file
	tmp, err := os.CreateTemp("", "traycli-icon-*.png")
	if err != nil {
		panic(err)
	}
	pngPath := tmp.Name()
	defer os.Remove(pngPath)
	if err := png.Encode(tmp, img); err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		panic(err)
	}

	// Output to project root (run from project root: go run ./scripts/genicon)
	wd, _ := os.Getwd()
	icoPath := filepath.Join(wd, "icon.ico")

	// Convert PNG to ICO
	ico := new(p2i.ICO).New(icoPath)
	ico.AddPngFile(pngPath)
	ico.Write()
}
