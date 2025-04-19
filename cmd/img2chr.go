package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"

	_ "image/png"
	_ "image/jpeg"
	_ "image/gif"

	"github.com/alexflint/go-arg"

	snesimg "github.com/zorchenhimer/go-snes/image"
)

type Arguments struct {
	Input  string `arg:"positional,required"`
	Output string `arg:"positional,required"`

	// Number of bits per pixel (colors per palette).
	// 1bpp=2, 2bpp=4, 4bpp=16, 8bpp=256, D=2047 max (maybe)
	// 1bpp is a special case meant for text.  This will have to be inflated to
	// 2bpp in the ROM software.
	BitDepth snesimg.BitDepth `arg:"--color-mode,-c" default:"2" help:"Bits per pixel. Accepted values are 1, 2, 4, & 8."`

	AsmOutput bool `arg:"--asm-out"`
}

func main() {
	args := &Arguments{}
	arg.MustParse(args)

	if err := run(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args *Arguments) error {
	input, err := os.Open(args.Input)
	if err != nil {
		return err
	}
	defer input.Close()

	img, _, err := image.Decode(input)
	if err != nil && errors.Is(err, image.ErrFormat) {
		return fmt.Errorf("CHR input not supported yet")
	} else if err != nil {
		return err
	}

	var pal color.Palette
	switch args.BitDepth {
	case snesimg.BD_1bpp:
		pal = snesimg.DefaultPal_1bpp
	case snesimg.BD_2bpp:
		pal = snesimg.DefaultPal_2bpp
	case snesimg.BD_4bpp:
		pal = snesimg.DefaultPal_4bpp
	case snesimg.BD_8bpp:
		pal = snesimg.DefaultPal_8bpp
	default:
		return fmt.Errorf("DirectColor not supported yet")
	}

	fmt.Println("BitDepth:", args.BitDepth)

	ti, err := snesimg.NewTiledImageFromImage(snesimg.CS_8x8, args.BitDepth, pal, img)
	if err != nil {
		return err
	}

	output, err := os.Create(args.Output)
	if err != nil {
		return err
	}
	defer output.Close()

	if args.AsmOutput {
		err = ti.WriteAsm(output)
	} else {
		err = ti.WriteBin(output)
	}

	return err
}
