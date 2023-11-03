package unifont

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	unifontHeight      = 16
	unifontNormalWidth = 8
	unifontWideWidth   = 16
)

type glyph struct {
	r      rune
	offset uint32
	wide   bool
}

type unifont struct {
	chardata       []byte
	glyphs         []glyph
	lastContinuous rune
	placeholder    *glyph
}

func parseStream(s io.Reader) (*unifont, error) {
	// should be large enough to not need to grow
	glyphs := make([]glyph, 0, 130000)
	chardata := bytes.NewBuffer(make([]byte, 0, 4*1024*1024))

	scanner := bufio.NewScanner(s)

	writetotal := 0
	lastRune := rune(-1)
	lastContinuous := rune(-1)
	placeholderOffset := -1

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) != 2 {
			return nil, errors.New("bad hex")
		}

		runeraw, err := strconv.ParseInt(parts[0], 16, 32)
		if err != nil {
			return nil, err
		}
		r := rune(runeraw)
		// sanity check, hex file should be sorted
		if r <= lastRune {
			return nil, errors.New("hex file not sorted")
		}
		lastRune = r

		if r == lastContinuous+1 {
			lastContinuous = r
		}

		charbits, err := hex.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}

		var wide bool
		if len(charbits) == 16 {
			wide = false
		} else if len(charbits) == 32 {
			wide = true
		} else {
			return nil, errors.New("bad character width")
		}

		_, err = chardata.Write(charbits)
		if err != nil {
			return nil, err
		}

		glyphs = append(glyphs, glyph{
			r:      r,
			offset: uint32(writetotal),
			wide:   wide,
		})

		if r == 0xFFFD {
			placeholderOffset = len(glyphs) - 1
		}

		writetotal += len(charbits)
	}

	// clip slices to reduce memory usage
	r := &unifont{
		chardata:       slices.Clip(chardata.Bytes()),
		glyphs:         slices.Clip(glyphs),
		lastContinuous: lastContinuous,
	}

	// save the placeholder glyph info
	if placeholderOffset >= 0 {
		r.placeholder = &r.glyphs[placeholderOffset]
	}

	return r, nil
}

// Creates a new "golang.org/x/image/font".Face object for the supplied Unifont .hex file
func NewFromHex(filename string) (*unifont, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseStream(f)
}

// Creates a new "golang.org/x/image/font".Face object for the supplied Unifont .hex.gz file
func NewFromHexGz(filename string) (*unifont, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return parseStream(gz)
}
