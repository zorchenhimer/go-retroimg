package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"

	snesimg "github.com/zorchenhimer/go-retroimg"
)

type Arguments struct {
	Input string `arg:"positional,required"`
	Config string `arg:"positional,required"`

	OutDir string `arg:"--output"`
}

func main() {
	args := &Arguments{}
	arg.MustParse(args)

	if err := run(args); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

type Segment struct {
	Start int
	Depth snesimg.BitDepth
	Count int
	Name string

	// In tiles, not pixels
	Width int
	Height int

	TileOrder  []int
	Sequential bool
}

func (s Segment) String() string {
	return fmt.Sprintf("{Segment Start:0x%X Count:0x%X (%d) Depth:%s TileOrder:%v}",
		s.Start,
		s.Count,
		s.Count,
		s.Depth,
		s.TileOrder,
	)
}

type CfgSegment struct {
	Start string
	Depth int
	Count string
	Name  string
	Dimensions string // WxH: 1x1, 2x1, 1x3, 2x2, etc
	TileOrder string // comma separated list of numbers
	Sequential bool
}

func parseConfig(filename string) ([]Segment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := []CfgSegment{}
	dec := json.NewDecoder(file)
	err = dec.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	segments := []Segment{}
	for _, seg := range cfg {
		start, err := strconv.ParseInt(seg.Start, 0, 32)
		if err != nil {
			return nil, err
		}

		count, err := strconv.ParseInt(seg.Count, 0, 32)
		if err != nil {
			return nil, err
		}

		w, h := 1, 1

		if seg.Dimensions != "" {
			if !strings.Contains(seg.Dimensions, "x") {
				return nil, fmt.Errorf("Invalid dimension: %q", seg.Dimensions)
			}

			dims := strings.Split(seg.Dimensions, "x")
			if len(dims) != 2 {
				return nil, fmt.Errorf("Invalid dimension: %q", seg.Dimensions)
			}

			w64, err := strconv.ParseInt(dims[0], 0, 32)
			if err != nil {
				return nil, fmt.Errorf("Invalid dimension: %q", seg.Dimensions)
			}

			h64, err := strconv.ParseInt(dims[1], 0, 32)
			if err != nil {
				return nil, fmt.Errorf("Invalid dimension: %q", seg.Dimensions)
			}

			w, h = int(w64), int(h64)
		}

		tileCount := w*h
		tileOrder := []int{}

		if seg.TileOrder == "" {
			for i := 0; i < tileCount; i++ {
				tileOrder = append(tileOrder, i+1)
			}
		} else {
			orderNums := strings.Split(seg.TileOrder, ",")
			highest := 0

			for _, str := range orderNums {
				n, err := strconv.Atoi(str)
				if err != nil {
					return nil, fmt.Errorf("Invalid TileOrder %q: %w", seg.TileOrder, err)
				}

				if n > highest {
					highest = n
				}
				tileOrder = append(tileOrder, n)
			}

			// FIXME: do i care about repeated numbers?
			if highest != len(tileOrder) {
				return nil, fmt.Errorf("Invalid TileOrder %q: bad length", seg.TileOrder)
			}
		}

		depth := snesimg.BD_2bpp
		if seg.Depth == 1 {
			depth = snesimg.BD_1bpp
		}

		if count < 1 {
			fmt.Println("Ignoring segment at", seg.Start)
			continue
		}

		seg.Name = strings.ReplaceAll(seg.Name, "{start}", seg.Start)
		seg.Name = strings.ReplaceAll(seg.Name, "{count}", seg.Count)
		seg.Name = strings.ReplaceAll(seg.Name, "{bpp}", strconv.Itoa(seg.Depth))

		segments = append(segments, Segment{
			Start: int(start),
			Depth: depth,
			Count: int(count),
			Name: seg.Name,
			Width: w,
			Height: h,
			TileOrder: tileOrder,
			Sequential: seg.Sequential,
		})
	}

	return segments, nil
}

func run(args *Arguments) error {
	segments, err := parseConfig(args.Config)
	if err != nil {
		return err
	}

	romfile, err := os.Open(args.Input)
	if err != nil {
		return err
	}
	defer romfile.Close()
	raw := snesimg.NewRawChr(romfile)

	if args.OutDir == "" {
		idx := strings.LastIndex(args.Input, ".")
		if idx < 1 {
			args.OutDir = args.Input+"_output"
		} else {
			args.OutDir = args.Input[:idx]
		}
	}

	err = os.MkdirAll(args.OutDir, 0755)
	if err != nil {
		return err
	}

	for num, seg := range segments {
		outname := fmt.Sprintf("%04d_%05X.png", num, seg.Start)
		if seg.Name != "" {
			outname = seg.Name
		}

		fmt.Println(outname, seg)

		_, err = romfile.Seek(int64(seg.Start), io.SeekStart)
		if err != nil {
			return fmt.Errorf("seek error (%05X): %w", seg.Start, err)
		}

		depth := snesimg.BitDepth(seg.Depth)
		pal, err := depth.DefaultPalette()
		if err != nil {
			return err
		}

		tilesPerTile := seg.Width * seg.Height

		MetaImage := &snesimg.MetaImage{
			Palette: pal,
			Stride: 16,
		}

		SEGLOOP:
		for i := 0; i < seg.Count; i++ {
			var tiles []*snesimg.Tile
			for j := 0; j < tilesPerTile; j++ {
				tile, err := raw.ReadTile(depth)
				if err != nil {
					if errors.Is(err, io.EOF) {
						fmt.Printf("found %d tiles\n", i)
						break SEGLOOP
					}
					return err
				}
				tiles = append(tiles, tile)
			}

			mt := snesimg.NewMetaTile(
				tiles,
				seg.Width,
				seg.Height,
				seg.TileOrder,
				MetaImage.Palette,
			)

			MetaImage.MetaTiles = append(MetaImage.MetaTiles, mt)
		}

		output, err := os.Create(filepath.Join(args.OutDir, outname))
		if err != nil {
			return err
		}

		err = png.Encode(output, MetaImage)
		output.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
