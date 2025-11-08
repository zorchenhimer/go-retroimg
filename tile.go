package retroimg

import (
	"image"
	"image/color"
	"fmt"
	"hash/crc32"
	"bytes"
)

// Tiles are always 8x8 pixels.
type Tile struct {
	image.Paletted

	Depth   BitDepth

	hash      string
	dirtyHash bool
}

func NewTile(depth BitDepth, palette color.Palette) *Tile {
	return &Tile{
		Paletted: image.Paletted{
			Pix:     make([]uint8, 64),
			Stride:  8,
			Rect:    image.Rect(0, 0, 8, 8),
			Palette: palette,
		},
		Depth: depth,

		dirtyHash: true,
	}
}

func NewTileFromPlanes(planes [][]byte) (*Tile, error) {
	var depth BitDepth

	switch len(planes) {
	case 1:
		depth = BD_1bpp
	case 2:
		depth = BD_2bpp
	case 4:
		depth = BD_4bpp
	case 8:
		depth = BD_8bpp
	default:
		return nil, fmt.Errorf("%d bit planes not supported", len(planes))
	}

	pal, _ := depth.DefaultPalette()
	tile := NewTile(depth, pal)

	for y := 0; y < 8; y++ {
		row := make([]byte, 8)

		for p := 0; p < len(planes); p++ {
			for x := 0; x < 8; x++ {
				bit := planes[p][y] & 0x01
				planes[p][y] = planes[p][y] >> 1
				//bit = bit << x
				row[7-x] = row[7-x] | bit << p
			}
		}

		for x := 0; x < 8; x++ {
			tile.Pix[(y*8)+x] = uint8(row[x])
		}

	}

	return tile, nil
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

// binary() returns all the bit planes as a binary slice.
// The number of bit planes is determined by Tile.Depth.
func (tile *Tile) binary() []byte {
	var numPlanes int
	switch tile.Depth {
	case BD_1bpp:
		numPlanes = 1
	case BD_2bpp:
		numPlanes = 2
	case BD_4bpp:
		numPlanes = 4
	case BD_8bpp:
		numPlanes = 8
	case BD_DirectColor:
		panic("DirectColor not implemented yet")
	default:
		panic("Unsupported bit depth")
	}

	planes := make([][]byte, numPlanes)
	for row := 0; row < 8; row++ {
		tmp := make([]byte, numPlanes)
		for col := 0; col < 8; col++ {
			color := tile.Pix[col+(row*8)]

			for plane := 0; plane < numPlanes; plane++ {
				tmp[plane] = tmp[plane] << 1 | (color & 1)
				color = color >> 1
			}
		}

		for i, t := range tmp {
			planes[i] = append(planes[i], t)
		}
	}

	return bytes.Join(planes, []byte{})
}
