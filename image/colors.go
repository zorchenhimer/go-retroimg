package image

import (
	//"fmt"
	"image/color"
	//"math"
)

var (
	DefaultPal_2bpp color.Palette
	DefaultPal_4bpp color.Palette
	DefaultPal_8bpp color.Palette
)

type BitDepth int

const (
	BD_2bpp BitDepth = iota
	BD_4bpp
	BD_8bpp
	BD_DirectColor
)

func init() {
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
