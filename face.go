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
func (f *Unifont) findGlyph(r rune) (*glyph, bool) {
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

func (f *Unifont) Close() error {
	// Because we read the entire font into memory, this is a no-op.
	return nil
}

func (f *Unifont) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	g, ok := f.findGlyph(r)
	// because we generate the glyph images at runtime
	if g != nil {
		x, y := dot.X.Round(), dot.Y.Round()
		if g.combining <= 0 {
			// combining character, offset the draw
			x += int(g.combining)
		} else {
			// non-combining character, advance dot
			advance = fixed.I(int(g.width))
		}
		dr = image.Rect(x, y, x+int(g.width), y+unifontHeight)
		mask = f.glyphImage(g)
		maskp = image.Point{}
	}
	return
}

func (f *Unifont) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	g, ok := f.findGlyph(r)
	if g != nil {
		xOffset := 0
		if g.combining <= 0 {
			// combining character, offset the bounding box
			xOffset = int(g.combining)
		} else {
			// non-combining character, advance dot
			advance = fixed.I(int(g.width))
		}
		bounds = fixed.R(xOffset, 0, xOffset+int(g.width), unifontHeight)
	}
	return
}

func (f *Unifont) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	g, ok := f.findGlyph(r)
	if g != nil {
		if g.combining > 0 {
			advance = fixed.I(int(g.width))
		}
	}
	return
}

func (f *Unifont) Kern(r0, r1 rune) fixed.Int26_6 {
	return 0
}

func (f *Unifont) Metrics() font.Metrics {
	return font.Metrics{
		Height:     fixed.I(unifontHeight),
		Ascent:     fixed.I(unifontHeight),
		Descent:    0,
		XHeight:    fixed.I(unifontHeight),
		CapHeight:  fixed.I(unifontHeight),
		CaretSlope: image.Pt(0, 1),
	}
}
