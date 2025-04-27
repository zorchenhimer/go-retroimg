package palette

import (
	"image/color"
	"strings"
	"io"
	"os"
	"errors"
	"fmt"
	"strconv"
	"bufio"

)

type ColorMap map[string]color.Color

func (cm ColorMap) FullPalette() color.Palette {
	pal := color.Palette{}
	for _, c := range cm {
		pal = append(pal, c)
	}
	return pal
}

func (cm ColorMap) NesPalette(C1, C2, C3, C4 string) color.Palette {
	c1, ok := cm[strings.ToLower(C1)]
	if !ok {
		c1 = cm["0f"]
	}

	c2, ok := cm[strings.ToLower(C2)]
	if !ok {
		c2 = cm["0f"]
	}

	c3, ok := cm[strings.ToLower(C3)]
	if !ok {
		c3 = cm["0f"]
	}

	c4, ok := cm[strings.ToLower(C4)]
	if !ok {
		c4 = cm["0f"]
	}

	return color.Palette{c1, c2, c3, c4}
}

type PaletteFormat int

const (
	// 8-bit red, green, blue in binary
	PF_RawRGB PaletteFormat = iota

	// Ascii text.  one color per line, RGB values in decimal delimited by tabs
	PF_Gimp
)

type PaletteDecodeFunc func(r io.Reader) (color.Palette, error)

func FromFile(filename string, format PaletteFormat) (color.Palette, error) {
	file, err := os.Open(filename)
	if err != nil {
		return color.Palette{}, err
	}
	defer file.Close()

	return FromReader(file, format)
}

func FromReader(r io.Reader, format PaletteFormat) (color.Palette, error) {
	var pal color.Palette
	var f PaletteDecodeFunc

	switch format {
	case PF_RawRGB:
		f = readRawRGB
	case PF_Gimp:
		f = readGimp
	default:
		return pal, fmt.Errorf("Unimplemnted format")
	}

	return f(r)
}

func readRawRGB(r io.Reader) (color.Palette, error) {
	var pal color.Palette
	var err error

	for {
		buf := make([]byte, 3)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			break
		}

		pal = append(pal, color.RGBA{uint8(buf[0]), uint8(buf[3]), uint8(buf[2]), 0xFF})
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return pal, err
}

func readGimp(r io.Reader) (color.Palette, error) {
	var pal color.Palette

	reader := bufio.NewScanner(r)
	first := true
	for reader.Scan() {
		line := reader.Text()
		if first {
			if strings.ToLower(line) != "gimp palette" {
				return pal, fmt.Errorf("missing 'GIMP Palette' on first line")
			}
			first = false
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			return pal, fmt.Errorf("Bad palette line: %q", line)
		}

		red, err := strconv.Atoi(parts[0])
		if err != nil {
			return pal, err
		}

		green, err := strconv.Atoi(parts[1])
		if err != nil {
			return pal, err
		}

		blue, err := strconv.Atoi(parts[2])
		if err != nil {
			return pal, err
		}

		pal = append(pal, color.RGBA{uint8(red), uint8(green), uint8(blue), 0xFF})
	}

	return pal, reader.Err()
}
