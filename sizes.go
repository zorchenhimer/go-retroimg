//go:build ignore

package retroimg

import (
	"image"
)

type Dimensions struct {
	// Tile counts in either direction.
	width int
	height int

	// Total image width in tiles (not bytes)
	stride int

	// Dimensions of the unerlying image format.
	// For NES and SNES, this is 8x8.
	dataWidth int
	dataHeight int

	// For dimensions with more than one row of tiles, do the subsequent
	// rows come immediately after the previous row's tiles, or after
	// [stride] number of tiles from the start of the previous row?
	sequential bool
}

var (
	Dim_8x16  = NewDimension(1, 2, 8, 8, 16, true)
	Dim_8x8   = NewDimension(1, 1, 8, 8, 16, true)
	Dim_16x16 = NewDimension(2, 2, 8, 8, 16, true)
)

var (
	Dim_NES = Dim_8x8
	Dim_NESSprite16 = Dim_8x16
)

func NewDimension(TileWidth, TileHeight, DataWidth, DataHeight, Stride int, Sequential bool) Dimensions {
	return Dimensions{
		width: Width,
		height: Height,
		datawidth: DataWidth,
		dataheight: DataHeight,
		stride: Stride,
		sequential: Sequential,
	}
}

func (d *Dimensions) Bounds() image.Rectangle {
	return image.Rect(0, 0, d.width, d.height)
}

// Offset into a list of tiles with this dimension.
func (d *Dimensions) TileOffset(x, y int) int {
	row := y / (d.height*d.dataHeight)
	col := x / (d.width*d.dataWidth)
	tx  := x % (d.width*d.dataWidth)
	ty  := y % (d.height*d.dataHeight)

	return row*32
}
