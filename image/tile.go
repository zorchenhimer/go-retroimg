package image

import (
	"image"
	"image/color"
	"fmt"
	"hash/crc32"
)

type BitDepth int

const (
	BD_2bpp BitDepth = iota
	BD_4bpp
	BD_8bpp
	BD_DirectColor
)

// Tiles are always 8x8 pixels.
type Tile struct {
	image.Paletted

	Id      int
	Depth   BitDepth

	hash      string
	dirtyHash bool
}

func NewTile(id int, depth BitDepth, palette color.Palette) *Tile {
	//depth := BD_DirectColor
	//switch len(color.Palette) {
	//case 4:
	//	depth = BD_2bpp
	//case 16:
	//	depth = BD_4bpp
	//case 256:
	//	depth = BD_8bpp
	//}

	return &Tile{
		Paletted: image.Paletted{
			Pix:     make([]uint8, 64),
			Stride:  8,
			Rect:    image.Rect(0, 0, 8, 8),
			Palette: palette,
		},
		Id:    id,
		Depth: depth,

		dirtyHash: true,
	}
}

func (this *Tile) IsIdentical(other *Tile) bool {
	if this.Hash() == other.Hash() {
		return true
	}
	return false
}

func (tile *Tile) Hash() string {
	if tile.dirtyHash || tile.hash == "" {
		tile.hash = fmt.Sprintf("%08X", crc32.ChecksumIEEE(tile.Pix))
	}

	return tile.hash
}

func (tile *Tile) At(x, y int) color.Color {
	// Return the "background" color if (x, y) is out of bounds.
	if 0 > y || y >= 8 || 0 > x || x >= 8 {
		return tile.Palette[0]
	}

	val := tile.Pix[(y*8)+x]
	return tile.Palette[val]
}

func (tile *Tile) Set(x, y int, c color.Color) {
	tile.dirtyHash = true
	tile.Paletted.Set(x, y, c)
}

func (tile *Tile) SetColorIndex(x, y int, idx uint8) {
	tile.dirtyHash = true
	tile.Paletted.SetColorIndex(x, y, idx)
}

func (tile *Tile) ColorModel() color.Model {
	return tile.Paletted.ColorModel()
}

func (tile *Tile) Bounds() image.Rectangle {
	return tile.Paletted.Rect
}
