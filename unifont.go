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

type glyph struct {
	r         rune
	offset    uint32
	width     uint8
	combining int8
}

// Unifont is a pared Unifont .hex file. It is safe to use concurrently with multiple font faces as
// long as no methods are called on it after being supplied to a font face.
type Unifont struct {
	chardata       []byte
	glyphs         []glyph
	lastContinuous rune
	placeholder    *glyph
}

// Optional flags to supply to the Unifont parsing functions
type UnifontOptions int

const (
	// Skip parsing characters in the Private Use Areas.
	NoPAUs UnifontOptions = iota
)

const (
	unifontHeight      = 16
	unifontNormalWidth = 8
	unifontWideWidth   = 16
)

// Parses the supplied Unifont .hex file from an input stream
func ParseHex(s io.Reader, options ...UnifontOptions) (*Unifont, error) {
	noPAUs := false

	for _, option := range options {
		switch option {
		case NoPAUs:
			noPAUs = true
		}
	}

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

		// skip PAU runes if skip flag is enabled
		if noPAUs &&
			((r >= 0xE000 && r <= 0xF8FF) ||
				(r >= 0xF0000 && r <= 0xFFFFF) ||
				(r >= 0x100000 && r <= 0x10FFFF)) {
			continue
		}

		if r == lastContinuous+1 {
			lastContinuous = r
		}

		charbits, err := hex.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}

		var width uint8
		if len(charbits) == 16 {
			width = unifontNormalWidth
		} else if len(charbits) == 32 {
			width = unifontWideWidth
		} else {
			return nil, errors.New("bad character width")
		}

		_, err = chardata.Write(charbits)
		if err != nil {
			return nil, err
		}

		glyphs = append(glyphs, glyph{
			r:         r,
			offset:    uint32(writetotal),
			width:     width,
			combining: 0x7F,
		})

		if r == 0xFFFD {
			placeholderOffset = len(glyphs) - 1
		}

		writetotal += len(charbits)
	}

	// clip slices to reduce memory usage
	r := &Unifont{
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

// Parses the supplied Unifont .hex file from a file
func ParseHexFile(filename string, options ...UnifontOptions) (*Unifont, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseHex(f, options...)
}

// Parses the supplied Unifont .hex file from a gzipped file
func ParseHexGzFile(filename string, options ...UnifontOptions) (*Unifont, error) {
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

	return ParseHex(gz, options...)
}
