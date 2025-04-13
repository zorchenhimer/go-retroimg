package main

import (
	"os"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/alexflint/go-arg"

	snesimg "github.com/zorchenhimer/go-snes/image"
)

type Arguments struct {
	CommandPng *CmdToPng   `arg:"subcommand:png" help:"Input is a CHR file.  Convert and output to PNG."`
	CommandChr *CmdToChr   `arg:"subcommand:chr" help:"Input is an image file.  Convert and output to CHR."`
	CommandPal *CmdPalette `arg:"subcommand:pal" help:"Input is a palette.  Convert to 15bit color suitable for the SNES"`

	//Input string `arg:"positional,required"`
	//Output string `arg:"positional,required"`

	// Number of bits per pixel (colors per palette).
	// 1bpp=2, 2bpp=4, 4bpp=16, 8bpp=256, D=2047 max (maybe)
	// 1bpp is a special case meant for text.  This will have to be inflated to
	// 2bpp in the ROM software.
	ColorMode string `arg:"--color-mode,-c" default:"2" help:"Bits per pixel. Accepted values are 2, 4, 8, and D (for Direct Color)."`

	// If provided, the input image will be converted to use this palette
	// before generating the output.  Number of colors must match ColorMode
	// (2bpp=4, 4bpp=16, 8bpp=256).  8bpp Direct Color is a special case that
	// will not require a specific
	// number of colors.
	PaletteFile string `arg:"--palette-file" help:"File containing a palette to use with the input image."`

	// TODO: make a writer that'll handle this?
	OutputFormat string `arg:"--output-format,-f" help:"Format of the output file.  Accepted values are 'bin' and 'asm' for binary and assembly (ca65 syntax), respectively."`
}

type CmdToPng struct {
	Input string `arg:"positional,required"`
	Output string `arg:"positional,required"`
}

type CmdToChr struct {
	Input string `arg:"positional,required"`
	Output string `arg:"positional,required"`

	// Remove duplicate and/or empty tiles from the input before writing
	// output.
	RemoveDuplicates bool `arg:"--rm-dupes" default:"false"`
	RemoveEmpty bool `arg:"--rm-empty" default:"false"`

	// The SNES can flip both sprite and nametable tiles horizontally and
	// vertically.  Implementing this will require outputting some sort of
	// metadata.
	RemoveFlipped bool `arg:"--rm-flipped" default:"false"`

	TileRemapFile string `arg:"--tile-remap" default:"File to write new tile IDs to when removing duplicate and/or empty tiles."`

	StartTileOffset int `arg:"--start-tile" default:"0" help:"Number of 8x8 tiles to skip from the input file before processing."`
	TileCount int `arg:"--tile-count" help:"Number of 8x8 tiles to process from the input file"`
	PadTiles  int `arg:"--pad-tiles" default:"0" help:"Pad output to have at least this many tiles.  Added tiles will be empty."`

	// The SNES supports tiles of different sizes.  Tiles destined for a
	// nametable can be 8x8, 16x16, and 16x8.  Tiles destined for sprites can
	// be 8x8, 16x16, 32x32, 64x64, 16x32, and 32x64.  This program does not
	// differentiate between sprite and nametable.
	TileSize string `arg:"--tile-size" help:"Size of the underlying tile.  Accepted values are 8x8, 16x16, 32x32, 64x64, 16x32, 32x64, and 16x8"`

	CharacterSize string `arg:"--char-size" default:"8x8"`
}

type CmdPalette struct {
	Input string `arg:"positional,required"`
	Output string `arg:"positional,required"`
}

func main() {
	args := &Arguments{}
	arg.MustParse(args)

	var err error
	switch {
	case args.CommandPng != nil:
		err = runCmdPng(args)
	case args.CommandChr != nil:
		err = runCmdChr(args)
	case args.CommandPal != nil:
	default:
		err = fmt.Errorf("Missing command")
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCmdPng(args *Arguments) error {
	infile, err := os.Open(args.CommandPng.Input)
	if err != nil {
		return err
	}
	defer infile.Close()

	pngImage, _, err := image.Decode(infile)
	if err != nil {
		return fmt.Errorf("DecodeConfig error: %w", err)
	}

	ti, err := snesimg.NewTiledImageFromImage(
		snesimg.CS_8x8, snesimg.BD_4bpp, snesimg.DefaultPal_4bpp, pngImage)
	if err != nil {
		return err
	}

	outfile, err := os.Create(args.CommandPng.Output)
	if err != nil {
		return err
	}
	defer outfile.Close()

	return png.Encode(outfile, ti)
}

func runCmdChr(args *Arguments) error {
	infile, err := os.Open(args.CommandChr.Input)
	if err != nil {
		return err
	}
	defer infile.Close()

	pngImage, _, err := image.Decode(infile)
	if err != nil {
		return fmt.Errorf("DecodeConfig error: %w", err)
	}

	var pal color.Palette
	switch pngImage.ColorModel().(type) {
	case color.Palette:
		fmt.Println("Paletted")
		pal = pngImage.ColorModel().(color.Palette)
		for _, c := range pal {
			r, g, b, _ := c.RGBA()
			r5, g5, b5 := uint8(r)/8, uint8(g)/8, uint8(b)/8
			all := uint16(r5&0x1F) | (uint16(g5&0x1F) << 5) | (uint16(b5&0x1F) << 10)

			fmt.Printf("  %02X %02X %02X -> $%04X (%%%04b_%04b %%%04b_%04b)\n",
				uint8(r), uint8(g), uint8(b), all,
				all>>12, (all>>8)&0x0F,
				(all>>4 & 0x0F), all & 0x0F)
		}
		fmt.Println("Colors:", len(pal))
	default:
		fmt.Println("[WARN] not an indexed image")
	}

	//fmt.Printf("%#v\n", cfg)
	fmt.Println("Bounds: ", pngImage.Bounds().Max)
	//fmt.Println("Height:", cfg.Height)

	var depth snesimg.BitDepth
	switch args.ColorMode {
	case "2":
		depth = snesimg.BD_2bpp
	case "4":
		depth = snesimg.BD_4bpp
	case "8":
		depth = snesimg.BD_8bpp
	case "D":
		return fmt.Errorf("no")
	}

	//infile.Seek(0, 0)
	//pngImage, _, err := image.Decode(infile)
	//if err != nil {
	//	return err
	//}

	var charSize snesimg.CharSize
	switch args.CommandChr.CharacterSize {
	case "8x8":
		charSize = snesimg.CS_8x8
	case "16x16":
		charSize = snesimg.CS_16x16
	case "16x8":
		charSize = snesimg.CS_16x8
	default:
		return fmt.Errorf("Invalid character size: %s", args.CommandChr.CharacterSize)
	}

	tm, err := snesimg.NewTilemapFromImage(charSize, depth, []color.Palette{pal}, pngImage)
	if err != nil {
		return fmt.Errorf("NewTilemap error: %w", err)
	}

	tmImage := tm.Image()
	outfile, err := os.Create(args.CommandChr.Output)
	if err != nil {
		return err
	}
	defer outfile.Close()

	err = png.Encode(outfile, tmImage)
	if err != nil {
		return err
	}

	return nil
}
