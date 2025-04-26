package image

import (
	"image/color"
	"io"
	"bufio"
	"strings"
	"strconv"
	"fmt"
)

func ReadTextPalettes(r io.Reader) (color.Palette, error) {
	pal := color.Palette{}
	reader := bufio.NewScanner(r)
	l := 1
	for reader.Scan() {
		line := reader.Text()
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "#$")

		// ignore short lines and comments
		if len(line) < 6 || strings.HasPrefix(line, ";") {
			continue
		}

		red, err := strconv.ParseUint(line[0:2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("red parse error on line %d: %w", l, err)
		}

		green, err := strconv.ParseUint(line[2:4], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("green parse error on line %d: %w", l, err)
		}

		blue, err := strconv.ParseUint(line[4:6], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("blue parse error on line %d: %w", l, err)
		}

		pal = append(pal, color.RGBA{uint8(red), uint8(green), uint8(blue), 0xFF})
		l++
	}

	return pal, reader.Err()
}
