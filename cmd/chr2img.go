package main

import (
	"fmt"
	"image/color"
	"image/png"
	"image/jpeg"
	"image/gif"
	"os"
	"strings"
	"path/filepath"

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
	BitDepth snesimg.BitDepth `arg:"--bit-depth,-d" default:"2" help:"Bits per pixel. Accepted values are 1, 2, 4, & 8 or 1bpp, 2bpp, 4bpp, & 8bpp."`

	PaletteFile string `arg:"--pal-file" help:"Read palette colors from this text file.  One color per line in HTML color syntax (eg #00AA55)."`
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
	var pal color.Palette
	var err error

	if args.PaletteFile != "" {
		palfile, err := os.Open(args.PaletteFile)
		if err != nil {
			return err
		}

		pal, err = snesimg.ReadTextPalettes(palfile)
		palfile.Close()
		if err != nil {
			return err
		}

	} else {
		pal, err = args.BitDepth.DefaultPalette()
		if err != nil {
			return err
		}
	}

	numColors, err := args.BitDepth.NumberColors()
	if err != nil {
		return err
	}


	if len(pal) < numColors {
		return fmt.Errorf("BitDepth of %s requires %d colors but palette only has %d", args.BitDepth, numColors, len(pal))
	} else if len(pal) > numColors {
		pal = pal[:numColors]
	}

	input, err := os.Open(args.Input)
	if err != nil {
		return err
	}
	defer input.Close()

	raw := snesimg.NewRawChr(input)
	tiles, err := raw.ReadAllTiles(args.BitDepth)
	if err != nil {
		return err
	}

	img := snesimg.NewTiledImageFromTiles(args.BitDepth, pal, tiles)

	output, err := os.Create(args.Output)
	if err != nil {
		return err
	}
	defer output.Close()

	switch strings.ToLower(filepath.Ext(args.Output)) {
	case ".png":
		err = png.Encode(output, img)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(output, img, &jpeg.Options{Quality: 90})
	case ".gif":
		err = gif.Encode(output, img, &gif.Options{NumColors: numColors})
	default:
		err = fmt.Errorf("Unsupported format")
	}

	return err
}
