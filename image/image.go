package image

import (
	"image"
	"fmt"
	"image/color"
)

var _ image.PalettedImage = &TiledImage{}

type TiledImage struct {
	Tiles   []TileMetadata
	TileIds []int

	// 8x8 or 16x16 "tiles".  Actual tiles are always 8x8, but 16x16 acts as a
	// metatile of sorts.
	CharacterSize CharSize

	Palette  color.Palette
	BitDepth BitDepth

	bounds image.Rectangle
}

func NewTiledImage(r image.Rectangle, cs CharSize, depth BitDepth, pal color.Palette) (*TiledImage, error) {
	switch cs {
	case CS_8x8, CS_16x16, CS_16x8, CS_32x32, CS_64x64, CS_16x32, CS_32x64:
		// valid
	default:
		return nil, fmt.Errorf("invalid CharSize: %#v", cs)
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
		return nil, fmt.Errorf("not implemented yet")
	}

	if len(pal) > palSize {
		return nil, fmt.Errorf("palette contains too many colors: %d; max: %d",
			len(pal), palSize)
	}

	width, height := cs.XY()
	if r.Max.X % width != 0 {
		return nil, fmt.Errorf("width of %d is not a multiple of %d", r.Max.X, width)
	}

	if r.Max.Y % height != 0 {
		return nil, fmt.Errorf("height of %d is not a multiple of %d", r.Max.Y, height)
	}

	mdSize := (r.Max.X/width)*(r.Max.Y/height)
	md := make([]TileMetadata, mdSize)
	ids := []int{}

	for i := 0; i < mdSize; i++ {
		md[i] = NewTileMetadata(cs, depth, pal)
		ids = append(ids, i)
	}


	return &TiledImage{
		Tiles:         md,
		CharacterSize: cs,
		Palette:       pal,
		BitDepth:      depth,

		bounds: r,
	}, nil
}

func NewTiledImageFromImage(cs CharSize, depth BitDepth, pal color.Palette, img image.Image) (*TiledImage, error) {
	ti, err := NewTiledImage(img.Bounds(), cs, depth, pal)
	if err != nil {
		return nil, err
	}

	var bppMod uint8 = 4
	if depth == BD_4bpp {
		fmt.Println("[BD_4bpp]")
		bppMod = 16
	}

	switch img.(type) {
	case *image.Paletted:
		fmt.Println("[Paletted]")
		palimg := img.(*image.Paletted)
		for y := 0; y < ti.bounds.Max.Y; y++ {
			for x := 0; x < ti.bounds.Max.X; x++ {
				if depth == BD_8bpp {
					ti.SetColorIndex(x, y, palimg.ColorIndexAt(x, y))
				} else {
					idx := palimg.ColorIndexAt(x, y) % bppMod
					ti.SetColorIndex(x, y, idx)
					//fmt.Printf("%3d ", idx)
				}
			}
			//fmt.Printf("\n")
		}

	default:
		fmt.Println("[RGB]")
		for y := 0; y < ti.bounds.Max.Y; y++ {
			for x := 0; x < ti.bounds.Max.X; x++ {
				ti.Set(x, y, img.At(x, y))
			}
		}
	}

	return ti, nil
}

func (ti *TiledImage) RemoveDuplicates() *TiledImage {
	panic("not yet")
}

func (ti *TiledImage) RemoveEmpty() *TiledImage {
	panic("not yet")
}

func (ti *TiledImage) Bounds() image.Rectangle {
	return ti.bounds
}

func (ti *TiledImage) Image() image.Image {
	img := image.NewRGBA(ti.Bounds())
	for y := 0; y < ti.bounds.Max.Y; y++ {
		for x := 0; x < ti.bounds.Max.X; x++ {
			img.Set(x, y, ti.At(x, y))
		}
	}
	return img
}

func (ti *TiledImage) At(x, y int) color.Color {
	width, height := ti.CharacterSize.XY()

	row := y / height
	col := x / width
	tx  := x % width
	ty  := y % height

	tileWidth := ti.bounds.Max.X/width

	return ti.Tiles[(row*tileWidth)+col].At(tx, ty)
}

func (ti *TiledImage) ColorIndexAt(x, y int) uint8 {
	width, height := ti.CharacterSize.XY()

	row := y / height
	col := x / width
	tx  := x % width
	ty  := y % height

	tileWidth := ti.bounds.Max.X/width

	return ti.Tiles[(row*tileWidth)+col].ColorIndexAt(tx, ty)
}

func (ti *TiledImage) Set(x, y int, c color.Color) {
	width, height := ti.CharacterSize.XY()

	row := y / height
	col := x / width
	tx  := x % width
	ty  := y % height

	tileWidth := ti.bounds.Max.X/width

	ti.Tiles[(row*tileWidth)+col].Set(tx, ty, c)
}

func (ti *TiledImage) SetColorIndex(x, y int, idx uint8) {
	width, height := ti.CharacterSize.XY()

	row := y / height
	col := x / width
	tx  := x % width
	ty  := y % height

	tileWidth := ti.bounds.Max.X/width

	ti.Tiles[(row*tileWidth)+col].SetColorIndex(tx, ty, idx)
}

func (ti *TiledImage) ColorModel() color.Model {
	return ti.Palette
}
