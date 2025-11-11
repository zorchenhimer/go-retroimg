package main

import (
	"errors"
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"

	_ "image/png"
	_ "image/jpeg"
	_ "image/gif"

	"github.com/alexflint/go-arg"

	snesimg "github.com/zorchenhimer/go-retroimg"
)

type Arguments struct {
	Input string `arg:"positional,required"`
	OutputBase string `arg:"positional,required"`

	// Number of bits per pixel (colors per palette).
	// 1bpp=2, 2bpp=4, 4bpp=16, 8bpp=256, D=2047 max (maybe)
	// 1bpp is a special case meant for text.  This will have to be inflated to
	// 2bpp in the ROM software.
	BitDepth snesimg.BitDepth `arg:"--bit-depth,-d" default:"2" help:"Bits per pixel. Accepted values are 1, 2, 4, & 8 or 1bpp, 2bpp, 4bpp, & 8bpp."`

	//AsmOutput bool `arg:"--asm-out"`
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

	pal, err := args.BitDepth.DefaultPalette()
	if err != nil {
		return err
	}

	fmt.Println("BitDepth:", args.BitDepth)

	ti, err := snesimg.NewTiledImageFromImage(snesimg.CS_8x8, args.BitDepth, pal, img)
	if err != nil {
		return err
	}

	if ti.Bounds().Max.X > 32*8 || ti.Bounds().Max.Y > 30*8 {
		return fmt.Errorf("Input image bounds too large: %#v", ti.Bounds().Max)
	}

	//ti.RemoveDuplicates()
	unique := ti.UniqueTiles()
	if len(unique) > 512 {
		return fmt.Errorf("Too many unique tiles: %d", len(unique))
	}

	if len(unique) > 256 {
		fmt.Println("WARN: unique tiles > 256 @", len(unique))
	}

	chrFile, err := os.Create(args.OutputBase+".chr")
	if err != nil {
		return err
	}

	err = unique.WriteChr(chrFile)
	if err != nil {
		return err
	}

	ntFile, err := os.Create(args.OutputBase+".nt.inc")
	if err != nil {
		return err
	}

	fmt.Fprintln(ntFile, ": .word", len(ti.TileIds))
	fmt.Fprint(ntFile, ": .byte ")

	ids := []string{}
	split := false
	for _, id := range ti.TileIds {
		if id > 255 && !split {
			fmt.Fprintln(ntFile, strings.Join(ids, ", "))
			fmt.Fprint(ntFile, ": .byte ")
			ids = []string{}
			split = true
		}
		ids = append(ids, strconv.Itoa(id&0xFF))
	}
	fmt.Fprintln(ntFile, strings.Join(ids, ", "))

	fmt.Println("len(ti.TileIds):", len(ti.TileIds))

	return nil
}

func main() {
	args := &Arguments{}
	arg.MustParse(args)

	if err := run(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

