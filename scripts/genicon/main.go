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
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	size     = 32
	text     = "CLI"
	greenHex = 0x00AA00
)

func main() {
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

	// Draw "CLI" centered with basicfont (approx 7x13 per char)
	// "CLI" = 3 chars * ~7 = 21 px wide, 13 px tall
	// Center: x = (32-21)/2 ≈ 5, y = (32+13)/2 ≈ 22 (baseline from top)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(green),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(5, 22),
	}
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
