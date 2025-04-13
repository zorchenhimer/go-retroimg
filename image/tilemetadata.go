package image

import (
	"image/color"
)

type TileMetadata struct {
	Tile8    *Tile   // 8x8px
	Tile16   []*Tile // 16x16px
	TileWide []*Tile // 16x8px

	FlipVertical   bool
	FlipHorizontal bool

	Palette    color.Palette
	PaletteIdx int
}

func NewTileMetadata(cs CharSize, depth BitDepth, pal color.Palette) TileMetadata {
	tm := TileMetadata{}
	switch cs {
	case CS_8x8:
		tm.Tile8 = NewTile(depth, pal)
	
	case CS_16x16:
		tm.Tile16 = make([]*Tile, 4)
		for i := 0; i < 4; i++ {
			tm.Tile16[i] = NewTile(depth, pal)
		}

	case CS_16x8:
		tm.TileWide = make([]*Tile, 2)
		for i := 0; i < 2; i++ {
			tm.TileWide[i] = NewTile(depth, pal)
		}
	default:
		panic("no")
	}

	return tm
}

func (tm *TileMetadata) At(x, y int) color.Color {
	if tm.Tile8 != nil {
		return tm.Tile8.At(x, y)
	}

	if tm.Tile16 != nil {
		row := y / 8
		col := x / 8
		return tm.Tile16[(row*2)+col].At(x%8, y%8)
	}

	if tm.TileWide != nil {
		//row := y / 8
		col := x / 8
		return tm.TileWide[col].At(x%8, y%8)
	}

	panic("no tile data in metatile")
}

func (tm *TileMetadata) Set(x, y int, c color.Color) {
	if tm.Tile8 != nil {
		tm.Tile8.Set(x, y, c)
		return
	}

	if tm.Tile16 != nil {
		row := y / 8
		col := x / 8
		tm.Tile16[(row*2)+col].Set(x%8, y%8, c)
		return
	}

	if tm.TileWide != nil {
		//row := y / 8
		col := x / 8
		tm.TileWide[col].Set(x%8, y%8, c)
		return
	}

	panic("no tile data in metatile")
}
