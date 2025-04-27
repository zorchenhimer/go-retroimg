package palettes

import (
	"image/color"
	"strings"
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
