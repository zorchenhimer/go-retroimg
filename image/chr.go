package image

import (
	"io"
	"errors"
)

type RawChr struct {
	r io.ReadSeeker
}

func NewRawChr(r io.ReadSeeker) *RawChr {
	return &RawChr{ r: r }
}

func (raw *RawChr) ReadTile(depth BitDepth) (*Tile, error) {
	planeCount, err := depth.PlaneCount()
	if err != nil {
		return nil, err
	}

	planes := [][]byte{}
	for i := 0; i < planeCount; i++ {
		buff := make([]byte, 8)
		n, err := io.ReadFull(raw.r, buff)
		if err != nil {

			if errors.Is(err, io.EOF) && i == 0 {
				return nil, io.EOF
			} else if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, err
		}

		if n != 8 {
			return nil, errors.New("Plane didn't read eight bytes")
		}

		planes = append(planes, buff)
	}

	return NewTileFromPlanes(planes)
}

func (raw *RawChr) ReadAllTiles(depth BitDepth) ([]*Tile, error) {
	tiles := []*Tile{}
	for {
		t, err := raw.ReadTile(depth)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return tiles, err
		}

		tiles = append(tiles, t)
	}

	return tiles, nil
}

func (raw *RawChr) DiscardTile(depth BitDepth) error {
	planeCount, err := depth.PlaneCount()
	if err != nil {
		return err
	}

	_, err = raw.r.Seek(int64(planeCount*8), io.SeekCurrent)
	return err
}

func (raw *RawChr) Seek(offset int64, whence int) (int64, error) {
	return raw.r.Seek(offset, whence)
}

func (raw *RawChr) Read(p []byte) (int, error) {
	return raw.r.Read(p)
}
