package unifont

import (
	"image"
	"slices"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Helper function for finding the glyph structure for a supplied rune. If the glyph is not found,
// the function returns false and the placeholder glyph (codepoint U+FFDF) if available. This is to
// match the expected behavior of the Glyph/GlyphBounds/GlyphAdvance interface functions.
func (f *unifont) findGlyph(r rune) (*glyph, bool) {
	if r < 0 {
		return f.placeholder, false
	}

	// in the continuous range, can do a direct slice offset lookup
	if r <= f.lastContinuous {
		return &f.glyphs[r], true
	}

	// outside the continuous range, do a binary search of the remaining glyphs
	i, ok := slices.BinarySearchFunc(f.glyphs[f.lastContinuous:], r, func(g glyph, r rune) int { return int(g.r - r) })
	if !ok {
		return f.placeholder, false
	}
	return &f.glyphs[int(f.lastContinuous)+i], true
}

func (f *unifont) Close() error {
	// Because we read the entire font into memory, this is a no-op.
	return nil
}

func (f *unifont) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	g, ok := f.findGlyph(r)
	// because we generate the glyph images at runtime
	if g != nil {
		x, y := dot.X.Round(), dot.Y.Round()
		var width int
		if g.wide {
			width = unifontWideWidth
		} else {
			width = unifontNormalWidth
		}
		dr = image.Rect(x, y, x+width, y+unifontHeight)
		mask = f.glyphImage(g)
		maskp = image.Point{}
		advance = fixed.I(width)
	}
	return
}

func (f *unifont) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	g, ok := f.findGlyph(r)
	if g != nil {
		if g.wide {
			bounds = fixed.R(0, 0, unifontWideWidth, unifontHeight)
			advance = fixed.I(unifontWideWidth)
		} else {
			bounds = fixed.R(0, 0, unifontNormalWidth, unifontHeight)
			advance = fixed.I(unifontNormalWidth)
		}
	}
	return
}

func (f *unifont) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	g, ok := f.findGlyph(r)
	if g != nil {
		if g.wide {
			advance = fixed.I(unifontWideWidth)
		} else {
			advance = fixed.I(unifontNormalWidth)
		}
	}
	return
}

func (f *unifont) Kern(r0, r1 rune) fixed.Int26_6 {
	// TODO: Maybe try to handle some combining characters?
	return fixed.I(0)
}

func (f *unifont) Metrics() font.Metrics {
	return font.Metrics{
		Height:     fixed.I(16),
		Ascent:     fixed.I(16),
		Descent:    fixed.I(0),
		XHeight:    fixed.I(16),
		CapHeight:  fixed.I(16),
		CaretSlope: image.Pt(0, 1),
	}
}
