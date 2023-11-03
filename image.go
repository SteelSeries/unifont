package unifont

import (
	"image"
	"image/color"
)

type unifontImage struct {
	pix  []byte
	wide bool
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
	if i.wide {
		return image.Rect(0, 0, unifontWideWidth, unifontHeight)
	} else {
		return image.Rect(0, 0, unifontNormalWidth, unifontHeight)
	}
}

func (i *unifontImage) At(x, y int) color.Color {
	if x < 0 || y < 0 {
		return color.Alpha16{}
	}
	if y >= unifontHeight {
		return color.Alpha16{}
	}
	if i.wide && x >= unifontWideWidth {
		return color.Alpha16{}
	}
	if !i.wide && x >= unifontNormalWidth {
		return color.Alpha16{}
	}

	var width int
	if i.wide {
		width = unifontWideWidth
	} else {
		width = unifontNormalWidth
	}
	offset := x + y*width
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
	var numPix uint32
	if g.wide {
		numPix = unifontHeight * unifontWideWidth
	} else {
		numPix = unifontHeight * unifontNormalWidth
	}
	return &unifontImage{
		pix:  f.chardata[g.offset : g.offset+numPix>>3],
		wide: g.wide,
	}
}
