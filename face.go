package unifont

import (
	"errors"
	"image"
	"slices"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type face struct {
	u *Unifont
	m int
}

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

func (f *face) Close() error {
	// Because we read the entire font into memory, this is a no-op.
	return nil
}

func (f *face) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	g, ok := f.u.findGlyph(r)
	// because we generate the glyph images at runtime
	if g != nil {
		yOffset := -unifontHeight * f.m
		x, y := dot.X.Round(), dot.Y.Round()
		if g.combining <= 0 {
			// combining character, offset the draw
			x += int(g.combining) * f.m
		} else {
			// non-combining character, advance dot
			advance = fixed.I(int(g.width) * f.m)
		}
		dr = image.Rect(x, y+yOffset, x+int(g.width)*f.m, y+unifontHeight*f.m+yOffset)
		mask = f.u.glyphImage(g, f.m)
		maskp = image.Point{}
	}
	return
}

func (f *face) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	g, ok := f.u.findGlyph(r)
	if g != nil {
		xOffset := 0
		yOffset := -unifontHeight * f.m
		if g.combining <= 0 {
			// combining character, offset the bounding box
			xOffset = int(g.combining) * f.m
		} else {
			// non-combining character, advance dot
			advance = fixed.I(int(g.width) * f.m)
		}
		bounds = fixed.R(xOffset, yOffset, int(g.width)*f.m+xOffset, unifontHeight*f.m+yOffset)
	}
	return
}

func (f *face) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	g, ok := f.u.findGlyph(r)
	if g != nil {
		if g.combining > 0 {
			advance = fixed.I(int(g.width) * f.m)
		}
	}
	return
}

func (f *face) Kern(r0, r1 rune) fixed.Int26_6 {
	return 0
}

func (f *face) Metrics() font.Metrics {
	return font.Metrics{
		Height:     fixed.I(unifontHeight * f.m),
		Ascent:     fixed.I(unifontHeight * f.m),
		Descent:    0,
		XHeight:    fixed.I(unifontHeight * f.m),
		CapHeight:  fixed.I(unifontHeight * f.m),
		CaretSlope: image.Pt(0, 1),
	}
}

// NewFace returns a new font.Face for the given Unifont. The multiplier parameter defines the
// output size of the font face. 1 means the default 16px height, 2 means 32px height, and so on.
func NewFace(u *Unifont, multipler int) (font.Face, error) {
	if multipler <= 0 {
		return nil, errors.New("multipler must be positive")
	}

	return &face{
		u: u,
		m: multipler,
	}, nil
}
