package main

import (
	"image"
	"image/png"
	"os"

	"github.com/ToadKing/unifont"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func main() {
	testString := "Hello World! 🙂🫵\uD7FF\uE000\U0001FFFF\U000F1C3F\U000FFFFE\U000FFFFF\U00100000\U0010FFFF\000"

	bg := image.NewRGBA(image.Rect(0, 0, 500, 300))
	fg := image.Black

	uf, err := unifont.NewFromHexGz("unifont_all-15.1.04.hex.gz")
	if err != nil {
		panic(err)
	}
	unifontDrawer := &font.Drawer{Src: fg, Dst: bg, Face: uf, Dot: fixed.P(30, 30)}
	unifontDrawer.DrawString(testString)

	gofont, err := opentype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	face, err := opentype.NewFace(gofont, nil)
	if err != nil {
		panic(err)
	}
	gofontDrawer := &font.Drawer{Src: fg, Dst: bg, Face: face, Dot: fixed.P(30, 230)}
	gofontDrawer.DrawString(testString)

	out, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	png.Encode(out, bg)
}
