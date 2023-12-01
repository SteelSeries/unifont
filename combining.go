package unifont

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Parse combining info from a combining.txt input stream.
//
// To get the data from the Unifont project to supply here, you can use this command on the command
// line in a copy of the Unicode source release:
//
//	shopt -s globstar
//	sort -u **/*-combining.txt > combining.txt
//
// This will generate a combined.txt file containing all the combining character offsets.
func (u *Unifont) CombiningInfo(s io.Reader) error {
	scanner := bufio.NewScanner(s)

	// To avoid modifying info in case of an error, save all the changes we'll do before we
	// actually apply them.
	type change struct {
		g         *glyph
		combining int8
	}
	// should be large enough to not need to grow
	changes := make(map[rune]change, 3000)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) != 2 {
			return errors.New("bad combining info")
		}

		runeraw, err := strconv.ParseInt(parts[0], 16, 32)
		if err != nil {
			return err
		}
		r := rune(runeraw)

		g, ok := u.findGlyph(r)
		if !ok {
			// skip nonexistent glyphs
			continue
		}

		if _, ok := changes[r]; ok {
			return fmt.Errorf("duplicate combining info on rune %x", r)
		}

		combining, err := strconv.ParseInt(parts[1], 10, 8)
		if err != nil {
			return err
		}

		changes[r] = change{g: g, combining: int8(combining)}
	}

	// actually apply changes now that no errors were detected
	for _, c := range changes {
		c.g.combining = c.combining
	}

	return nil
}

// Parse combining info from a combining.txt file.
func (u *Unifont) CombiningInfoFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return u.CombiningInfo(f)
}

// Parse combining info from a combining.txt.gz stream.
func (u *Unifont) CombiningInfoGzFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	return u.CombiningInfo(gz)
}
