package main

import (
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Tileset interface {
	Image(id int) *ebiten.Image
}

type UniformTilesetJSON struct {
	Path string `json:"image"`
}

type UniformTileset struct {
	image *ebiten.Image
	gid   int
}

func (u *UniformTileset) Image(id int) *ebiten.Image {
	id -= u.gid

	// Get the position on the image where the title id is
	srcX := id % 22
	srcY := id / 22

	// Convert the src tile pos to pixel src position
	srcX *= 16
	srcY *= 16

	return u.image.SubImage(
		image.Rect(
			srcX, srcY,
			srcX+16, srcY+16,
		),
	).(*ebiten.Image)
}

type TileJSON struct {
	Id     int    `json:"id"`
	Path   string `json:"image"`
	Width  int    `json:"imagewidth"`
	Height int    `json:"imageheight"`
}

type DynamicTilesetJSON struct {
	Tiles []*TileJSON `json:"tiles"`
}

type DynamicTileset struct {
	images []*ebiten.Image
	gid    int
}

func (d *DynamicTileset) Image(id int) *ebiten.Image {
	id -= d.gid

	return d.images[id]
}

func NewTileset(path string, gid int) (Tileset, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if strings.Contains(path, "building") {
		// return dynamic tileset
		var dynamicTilesetJson DynamicTilesetJSON
		err = json.Unmarshal(contents, &dynamicTilesetJson)
		if err != nil {
			return nil, err
		}

		dynamicTileset := DynamicTileset{}
		dynamicTileset.gid = gid
		dynamicTileset.images = make([]*ebiten.Image, 0)

		for _, tileJSON := range dynamicTilesetJson.Tiles {
			tileJSONPath := tileJSON.Path
			tileJSONPath = filepath.Clean(tileJSONPath)
			tileJSONPath = strings.ReplaceAll(tileJSONPath, "\\", "/")
			tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")
			tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")
			tileJSONPath = filepath.Join("assets/", tileJSONPath)

			img, _, err := ebitenutil.NewImageFromFile(tileJSONPath)
			if err != nil {
				return nil, err
			}

			dynamicTileset.images = append(dynamicTileset.images, img)
		}
		return &dynamicTileset, nil
	}

	// return uniform tileset
	var uniformTilesetJSON UniformTilesetJSON
	err = json.Unmarshal(contents, &uniformTilesetJSON)
	if err != nil {
		return nil, err
	}

	uniformTileset := UniformTileset{}

	tileJSONPath := uniformTilesetJSON.Path
	tileJSONPath = filepath.Clean(tileJSONPath)
	tileJSONPath = strings.ReplaceAll(tileJSONPath, "\\", "/")
	tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")
	tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")
	tileJSONPath = filepath.Join("assets/", tileJSONPath)

	img, _, err := ebitenutil.NewImageFromFile(tileJSONPath)
	if err != nil {
		return nil, err
	}
	uniformTileset.image = img
	uniformTileset.gid = gid

	return &uniformTileset, nil

}
