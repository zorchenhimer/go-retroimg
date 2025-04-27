package retroimg

import (
	"image"
	"image/color"
	"io"
	"fmt"
)

type CharSize int

const (
	CS_8x8 CharSize = iota
	CS_16x16
	CS_16x8
	CS_32x32
	CS_64x64
	CS_16x32
	CS_32x64
)

func (cs CharSize) XY() (int, int) {
	switch cs {
	case CS_8x8:
		return 8, 8
	case CS_16x16:
		return 16, 16
	case CS_16x8:
		return 16, 8
	case CS_32x32:
		return 32, 32
	case CS_64x64:
		return 64, 64
	case CS_16x32:
		return 16, 32
	case CS_32x64:
		return 32, 64
	}

	panic(fmt.Sprintf("invalid CharSize: %d", int(cs)))
}

type Tilemap struct {
	Tiles []TileMetadata

	// 8x8 or 16x16 "tiles".  Actual tiles are always 8x8, but 16x16 acts as a
	// metatile of sorts.
	CharacterSize CharSize

	Palettes []color.Palette
}

func validateTilemapValues(cs CharSize, depth BitDepth, pals []color.Palette) error {
	switch cs {
	case CS_8x8, CS_16x16, CS_16x8:
		// valid
	default:
		return fmt.Errorf("invalid CharSize: %#v", cs)
	}

	if len(pals) > 8 {
		return fmt.Errorf("too many palettes")
	} else if len(pals) <= 0 {
		return fmt.Errorf("too few palettes")
	}

	// Max value
	var palSize int

	switch depth {
	case BD_2bpp:
		palSize = 4
	case BD_4bpp:
		palSize = 16
	case BD_8bpp:
		palSize = 256
	case BD_DirectColor:
		return fmt.Errorf("not implemented yet")
	}

	for i, p := range pals {
		if len(p) > palSize { // DirectColor??
			return fmt.Errorf("palette at index %d contains too many colors: %d; max: %d",
				i, len(p), palSize)
		}
	}

	return nil
}

func NewTilemap(cs CharSize, depth BitDepth, pals []color.Palette) (*Tilemap, error) {
	err := validateTilemapValues(cs, depth, pals)
	if err != nil {
		return nil, err
	}

	md := make([]TileMetadata, 32*32)

	for i := 0; i < 32*32; i++ {
		md[i] = NewTileMetadata(cs, depth, pals[0])
	}

	return &Tilemap{
		Tiles: md,
	}, nil
}

func NewTilemapFromImage(cs CharSize, depth BitDepth, pals []color.Palette, img image.Image) (*Tilemap, error) {
	err := validateTilemapValues(cs, depth, pals)
	if err != nil {
		return nil, err
	}

	md := make([]TileMetadata, 32*32)

	for i := 0; i < 32*32; i++ {
		md[i] = NewTileMetadata(cs, depth, pals[0])
	}

	tm := &Tilemap{
		Tiles: md,
		CharacterSize: cs,
		Palettes: pals,
	}

	sizeX, sizeY := cs.XY()
	for y := 0; y < 32*sizeY; y++ {
		for x := 0; x < 32*sizeX; x++ {
			tm.Set(x, y, img.At(x, y))
		}
	}

	return tm, nil
}

func (tm *Tilemap) Bounds() image.Rectangle {
	x, y := tm.CharacterSize.XY()
	return image.Rect(0, 0, 32*x, 32*y)
}

func (tm *Tilemap) Image() image.Image {
	img := image.NewRGBA(tm.Bounds())
	for y := 0; y < tm.Bounds().Max.Y; y++ {
		for x := 0; x < tm.Bounds().Max.X; x++ {
			img.Set(x, y, tm.At(x, y))
		}
	}
	return img
}

func (tm *Tilemap) ChrBin() []byte {
	panic("not yet")
}

func (tm *Tilemap) ChrAsm() []string {
	panic("not yet")
}

func (tm *Tilemap) WriteChr(w io.Writer) error {
	panic("not yet")
}

func (tm *Tilemap) WriteAsm(w io.Writer) error {
	panic("not yet")
}

func (tm *Tilemap) At(x, y int) color.Color {
	width, height := tm.CharacterSize.XY()

	row := y / height
	col := x / width
	tx  := x % width
	ty  := y % height

	return tm.Tiles[(row*32)+col].At(tx, ty)
}

func (tm *Tilemap) Set(x, y int, c color.Color) {
	width, height := tm.CharacterSize.XY()

	row := y / height
	col := x / width
	tx  := x % width
	ty  := y % height

	tm.Tiles[(row*32)+col].Set(tx, ty, c)
}
