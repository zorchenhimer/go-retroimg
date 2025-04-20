package image

import (
	"fmt"
	"image/color"
	"strings"
	//"math"
)

var (
	DefaultPal_1bpp color.Palette
	DefaultPal_2bpp color.Palette
	DefaultPal_4bpp color.Palette
	DefaultPal_8bpp color.Palette
)

type BitDepth int

const (
	BD_1bpp BitDepth = iota
	BD_2bpp
	BD_4bpp
	BD_8bpp
	BD_DirectColor
)

func (bd *BitDepth) UnmarshalText(b []byte) error {
	switch strings.ToLower(strings.TrimSpace(string(b))) {
	case "1", "1bpp":
		*bd = BD_1bpp
	case "2", "2bpp":
		*bd = BD_2bpp
	case "4", "4bpp":
		*bd = BD_4bpp
	case "8", "8bpp":
		*bd = BD_8bpp
	case "d", "direct", "directcolor":
		*bd = BD_DirectColor
	default:
		return fmt.Errorf("Invalid bit depth value: %q", string(b))
	}

	return nil
}

func (bd BitDepth) String() string {
	switch bd {
	case BD_1bpp:
		return "BD_1bpp"
	case BD_2bpp:
		return "BD_2bpp"
	case BD_4bpp:
		return "BD_4bpp"
	case BD_8bpp:
		return "BD_8bpp"
	case BD_DirectColor:
		return "BD_DirectColor"
	default:
		return "UNKNOWN"
	}
}

func init() {
	DefaultPal_1bpp = color.Palette{
		color.Gray{0x00},
		color.Gray{0xFF},
	}

	DefaultPal_2bpp = color.Palette{
		color.Gray{0x00},
		color.Gray{0x55},
		color.Gray{0xAA},
		color.Gray{0xFF},
	}

	DefaultPal_8bpp = color.Palette{}
	for i := 0; i < 256; i++ {
		DefaultPal_8bpp = append(DefaultPal_8bpp, color.Gray{uint8(i)})
	}

	DefaultPal_4bpp = color.Palette{}
	for i := 0; i < 16; i++ {
		c := color.Gray{ uint8(i << 4) }
		DefaultPal_4bpp = append(DefaultPal_4bpp, c)
		//fmt.Printf("%3d ", c.Y)
	}
	//fmt.Printf("\n")

	//for _, p := range DefaultPal_4bpp {
	//	g := p.(color.Gray)
	//	fmt.Printf("0x%02X\n", g.Y)
	//}
}
