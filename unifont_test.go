package unifont

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Test for general functionality
func TestGeneral(t *testing.T) {
	const testString = "He\u0300ll\u08F2o World! 🫵🙂\uD7FF\uE000\U0001FFFF\U000F1C3F\U000FFFFE\U000FFFFF\U00100000\U0010FFFF\000"

	bg := image.NewRGBA(image.Rect(0, 0, 1500, 300))
	draw.Draw(bg, bg.Bounds(), image.White, image.Point{}, draw.Src)
	fg := image.Black

	uf, err := ParseHexGzFile("testdata/unifont_all-15.1.04.hex.gz")
	if err != nil {
		panic(err)
	}
	err = uf.CombiningInfoGzFile("testdata/combining.txt.gz")
	if err != nil {
		panic(err)
	}

	// test concurrency
	var wg sync.WaitGroup
	for _, x := range []struct {
		y int
		m int
	}{{20, 1}, {120, 6}} {
		wg.Add(1)
		go func(y, m int) {
			face, err := NewFace(uf, m)
			if err != nil {
				panic(err)
			}
			var wg2 sync.WaitGroup
			for _, y2 := range []int{10, 160} {
				wg2.Add(1)
				go func(face font.Face, y, m int) {
					drawer := &font.Drawer{Src: fg, Dst: bg, Face: face, Dot: fixed.P(10, y)}
					drawer.DrawString(testString)
					wg2.Done()
				}(face, y+y2, m)
			}
			wg2.Wait()
			wg.Done()
		}(x.y, x.m)
	}
	wg.Wait()

	// code for writing new PNGs for updated test results
	// out, err := os.Create("out.png")
	// if err != nil {
	// 	panic(err)
	// }
	// defer out.Close()
	// err = png.Encode(out, bg)
	// if err != nil {
	// 	panic(err)
	// }

	expectedFile, err := os.Open("testdata/expected.png")
	if err != nil {
		panic(err)
	}
	defer expectedFile.Close()

	expected, err := png.Decode(expectedFile)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, expected, bg)
}
