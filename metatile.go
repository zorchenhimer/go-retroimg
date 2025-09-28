package retroimg

import (
	"image"
	"image/color"
)

type MetaImage struct {
	MetaTiles []*MetaTile
	Palette color.Palette
	Stride int
}

func (mi *MetaImage) ColorModel() color.Model {
	return mi.Palette
}

func (mi *MetaImage) Bounds() image.Rectangle {
	mtBounds := mi.MetaTiles[0].Bounds()
	return image.Rect(0, 0,
		mtBounds.Max.X * mi.Stride,
		mtBounds.Max.Y * ((len(mi.MetaTiles)+mi.Stride-1) / mi.Stride),
	)
}

func (mi *MetaImage) At(x, y int) color.Color {
	rect := mi.Bounds()
	if x < rect.Min.X || y < rect.Min.Y || x >= rect.Max.X || y >= rect.Max.Y {
		return mi.Palette[0]
	}

	mtBounds := mi.MetaTiles[0].Bounds()

	row := y / mtBounds.Max.Y
	col := x / mtBounds.Max.X
	tx  := x % mtBounds.Max.X
	ty  := y % mtBounds.Max.Y
	idx := (row*mi.Stride)+col

	if idx >= len(mi.MetaTiles) {
		return mi.Palette[0]
	}

	return mi.MetaTiles[idx].At(tx, ty)
}

type MetaTile struct {
	Tiles []*Tile
	OrderedTiles []*Tile

	// In tiles, not pixels
	Width int
	Height int

	TileOrder []int
	Palette color.Palette
}

func NewMetaTile(tiles []*Tile, width, height int, tileOrder []int, pal color.Palette) *MetaTile {
	ordered := []*Tile{}
	for _, id := range tileOrder {
		ordered = append(ordered, tiles[id-1])
	}

	return &MetaTile{
		Tiles: tiles,
		Width: width,
		Height: height,
		TileOrder: tileOrder,
		OrderedTiles: ordered,
		Palette: tiles[0].Palette,
	}
}

func (mt *MetaTile) At(x, y int) color.Color {
	rect := mt.Bounds()
	if x < rect.Min.X || y < rect.Min.Y || x >= rect.Max.X || y >= rect.Max.Y {
		return mt.Palette[0]
	}

	tileBounds := mt.Tiles[0].Bounds()

	row := y / tileBounds.Max.Y
	col := x / tileBounds.Max.X
	tx  := x % tileBounds.Max.X
	ty  := y % tileBounds.Max.Y

	return mt.OrderedTiles[(row*mt.Width)+col].At(tx, ty)
}

func (mt *MetaTile) ColorModel() color.Model {
	return mt.Palette
}

func (mt *MetaTile) Bounds() image.Rectangle {
	tileBounds := mt.Tiles[0].Bounds()
	return image.Rect(0, 0,
		tileBounds.Max.X * mt.Width,
		tileBounds.Max.Y * mt.Height,
	)
}

