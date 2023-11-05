package unifont

import (
	"image"
	"image/color"
)

type unifontImage struct {
	pix   []byte
	width uint8
}

func monochromeAlphaConversion(c color.Color) color.Color {
	if _, _, _, a := c.RGBA(); a == 0xFFFF {
		return color.Alpha16{A: 0xFFFF}
	} else {
		return color.Alpha16{A: 0}
	}
}

var (
	monochromeAlphaModel = color.ModelFunc(monochromeAlphaConversion)
)

func (i *unifontImage) ColorModel() color.Model {
	return monochromeAlphaModel
}

func (i *unifontImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(i.width), unifontHeight)
}

func (i *unifontImage) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= int(i.width) || y >= unifontHeight {
		return color.Alpha16{}
	}

	offset := x + y*int(i.width)
	byteOffset := offset >> 3
	bitOffset := 7 - (offset & 7)
	val := i.pix[byteOffset] & (1 << bitOffset)
	if val != 0 {
		return color.Alpha16{A: 0xFFFF}
	} else {
		return color.Alpha16{A: 0}
	}
}

func (f *unifont) glyphImage(g *glyph) *unifontImage {
	return &unifontImage{
		pix:   f.chardata[g.offset : g.offset+uint32(g.width)*unifontHeight>>3],
		width: g.width,
	}
}
