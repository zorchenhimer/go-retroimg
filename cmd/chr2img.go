package main

import (
	"fmt"
	"io"
	"image/color"
	"image/png"
	"image/jpeg"
	"image/gif"
	"os"
	"strings"
	"strconv"
	"path/filepath"

	"github.com/alexflint/go-arg"

	snesimg "github.com/zorchenhimer/go-retroimg"
	"github.com/zorchenhimer/go-retroimg/palette"
)

type Arguments struct {
	Input  string `arg:"positional,required"`
	Output string `arg:"positional,required"`

	// Number of bits per pixel (colors per palette).
	// 1bpp=2, 2bpp=4, 4bpp=16, 8bpp=256, D=2047 max (maybe)
	// 1bpp is a special case meant for text.  This will have to be inflated to
	// 2bpp in the ROM software.
	BitDepth snesimg.BitDepth `arg:"--bit-depth,-d" default:"2" help:"Bits per pixel. Accepted values are 1, 2, 4, & 8 or 1bpp, 2bpp, 4bpp, & 8bpp."`

	// --nes-pal 0F,00,1A,20
	NesPal string `arg:"--nes-pal"`

	PaletteFile string `arg:"--pal-file" help:"Read palette colors from this text file.  One color per line in HTML color syntax (eg #00AA55)."`

	StartOffset string `arg:"--start"`
	startOffset int
	TileCount string `arg:"--tile-count"`
	tileCount int
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

	if args.PaletteFile != "" && args.NesPal != "" {
		return fmt.Errorf("Cannot use both --nes-pal and --pal-file")
	}

	if args.PaletteFile != "" {
		pal, err = palette.FromFile(args.PaletteFile, palette.PF_Gimp)
		if err != nil {
			return err
		}

	} else if args.NesPal != "" {
		if args.BitDepth != snesimg.BD_2bpp {
			return fmt.Errorf("Can only use --nes-pal with a 2bpp image")
		}

		parts := strings.Split(args.NesPal, ",")
		if len(parts) < 4 {
			return fmt.Errorf("Too few colors")
		}
		if len(parts) > 4 {
			return fmt.Errorf("Too many colors")
		}

		for i := 0; i < len(parts); i++ {
			parts[i] = strings.TrimLeft(parts[i], "$")
		}

		pal = palette.Nes_2C02.NesPalette(parts[0], parts[1], parts[2], parts[3])
		fmt.Println(pal)

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

	if args.StartOffset != "" {
		offset, err := strconv.ParseInt(args.StartOffset, 0, 32)
		if err != nil {
			return err
		}

		_, err = input.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("Seek() error: %w", err)
		}
	}

	raw := snesimg.NewRawChr(input)

	var tiles []*snesimg.Tile
	if args.TileCount != "" {
		count, err := strconv.ParseInt(args.TileCount, 0, 32)
		if err != nil {
			return err
		}

		fmt.Println("count:", count)

		for i := 0; i < int(count); i++ {
			t, err := raw.ReadTile(args.BitDepth)
			if err != nil {
				//return err
				fmt.Printf("read tile err: %s\n", err)
				break
			}
			tiles = append(tiles, t)
		}
	} else {
		tiles, err = raw.ReadAllTiles(args.BitDepth)
		if err != nil {
			return err
		}
	}

	img := snesimg.NewTiledImageFromTiles(args.BitDepth, pal, tiles)
	fmt.Printf("Bounds: %#v\n", img.Bounds())
	fmt.Println("len(tiles):", len(tiles))

	args.Output = strings.ReplaceAll(args.Output, "{start}", args.StartOffset)
	args.Output = strings.ReplaceAll(args.Output, "{count}", args.TileCount)
	args.Output = strings.ReplaceAll(args.Output, "{bpp}", args.BitDepth.String())
	fmt.Println("output:", args.Output)

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
