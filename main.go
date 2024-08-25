package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"rpg-tutorial/entities"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	player       *entities.Player
	enemies      []*entities.Enemy
	potions      []*entities.Potion
	tilemapJSON  *TilemapJSON
	tilesets     []Tileset
	tilemapImage *ebiten.Image
	camera       *Camera
	colliders    []image.Rectangle
}

func CheckCollisionsHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+16,
				int(sprite.Y)+16),
		) {
			if sprite.Vx > 0.0 {
				sprite.X = float64(collider.Min.X) - 16.0
			} else if sprite.Vx < 0.0 {
				sprite.X = float64(collider.Max.X)
			}
		}
	}
}

func CheckCollisionsVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+16,
				int(sprite.Y)+16),
		) {
			if sprite.Vy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - 16.0
			} else if sprite.Vy < 0.0 {
				sprite.Y = float64(collider.Max.Y)
			}
		}
	}
}

func (g *Game) Update() error {
	// React to keypressed

	g.player.Vx = 0
	g.player.Vy = 0
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Vx += 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Vx -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Vy -= 2

	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Vy += 2
	}

	g.player.X += g.player.Vx

	CheckCollisionsHorizontal(g.player.Sprite, g.colliders)

	g.player.Y += g.player.Vy

	CheckCollisionsVertical(g.player.Sprite, g.colliders)

	for _, sprite := range g.enemies {
		sprite.Vx = 0
		sprite.Vy = 0
		if sprite.FollowsPlayer {
			if sprite.X < g.player.X {
				sprite.Vx += 1
			} else if sprite.X > g.player.X {
				sprite.Vx -= 1
			}
			if sprite.Y < g.player.Y {
				sprite.Vy += 1
			} else if sprite.Y > g.player.Y {
				sprite.Vy -= 1
			}
		}

		sprite.X += sprite.Vx

		CheckCollisionsHorizontal(sprite.Sprite, g.colliders)

		sprite.Y += sprite.Vy

		CheckCollisionsVertical(sprite.Sprite, g.colliders)
	}

	for _, potion := range g.potions {
		if g.player.X > potion.X {
			g.player.Health += potion.AmountHeal
			fmt.Printf("Picked up potion. Health: %d\n", g.player.Health)
		}
	}

	g.camera.FollowTarget(g.player.X+8, g.player.Y+8, 320.0, 240.0)
	g.camera.Constrain(
		float64(g.tilemapJSON.Layers[0].Width)*16.0,
		float64(g.tilemapJSON.Layers[0].Height)*16.0,
		320,
		240,
	)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{120, 180, 255, 255})

	options := ebiten.DrawImageOptions{}

	// Loop over the layers
	for layerIndex, layer := range g.tilemapJSON.Layers {
		for index, id := range layer.Data {

			if id == 0 {
				continue
			}

			x := index % layer.Width
			y := index / layer.Width

			x *= 16
			y *= 16

			image := g.tilesets[layerIndex].Image(id)

			options.GeoM.Translate(float64(x), float64(y))

			options.GeoM.Translate(0.0, -(float64(image.Bounds().Dy()) + 16))

			options.GeoM.Translate(g.camera.X, g.camera.Y)

			screen.DrawImage(image, &options)

			options.GeoM.Reset()
		}
	}

	options.GeoM.Translate(g.player.X, g.player.Y)
	options.GeoM.Translate(g.camera.X, g.camera.Y)

	// Draw the player
	screen.DrawImage(
		g.player.Image.SubImage(
			image.Rect(0, 0, 16, 16),
		).(*ebiten.Image),
		&options,
	)
	options.GeoM.Reset()

	for _, sprite := range g.enemies {
		options.GeoM.Translate(sprite.X, sprite.Y)
		options.GeoM.Translate(g.camera.X, g.camera.Y)

		screen.DrawImage(
			sprite.Image.SubImage(
				image.Rect(0, 0, 16, 16),
			).(*ebiten.Image),
			&options,
		)

		options.GeoM.Reset()
	}

	for _, sprite := range g.potions {
		options.GeoM.Translate(sprite.X, sprite.Y)
		options.GeoM.Translate(g.camera.X, g.camera.Y)

		screen.DrawImage(
			sprite.Image.SubImage(
				image.Rect(0, 0, 16, 16),
			).(*ebiten.Image),
			&options,
		)

		options.GeoM.Reset()
	}

	for _, collider := range g.colliders {
		vector.StrokeRect(
			screen,
			float32(collider.Min.X)+float32(g.camera.X),
			float32(collider.Min.Y)+float32(g.camera.Y),
			float32(collider.Dx()),
			float32(collider.Dy()),
			1.0,
			color.RGBA{255, 0, 0, 255},
			true,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	playerImage, _, err := ebitenutil.NewImageFromFile("./assets/images/ninja.png")
	if err != nil {
		log.Fatal(err)
	}

	skeletonImage, _, err := ebitenutil.NewImageFromFile("./assets/images/skeleton.png")
	if err != nil {
		log.Fatal()
	}

	potionImage, _, err := ebitenutil.NewImageFromFile("./assets/images/potion.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapJSON, err := NewTilemapJSON("assets/maps/spawn.json")
	if err != nil {
		log.Fatal(err)
	}

	tilesets, err := tilemapJSON.GetTilesets()
	if err != nil {
		log.Fatal(err)
	}

	tilemapImage, _, err := ebitenutil.NewImageFromFile("./assets/images/TilesetFloor.png")
	if err != nil {
		log.Fatal(err)
	}

	game := Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Image: playerImage,
				X:     50.0,
				Y:     50.0,
			},
			Health: 10,
		},
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Image: skeletonImage,
					X:     100.0,
					Y:     100.0,
				},
				FollowsPlayer: false,
			},
			{
				Sprite: &entities.Sprite{
					Image: skeletonImage,
					X:     150.0,
					Y:     100.0,
				},
				FollowsPlayer: true,
			},
		},
		potions: []*entities.Potion{
			{
				Sprite: &entities.Sprite{
					Image: potionImage,
					X:     250,
					Y:     90,
				},
				AmountHeal: 1,
			},
		},
		tilemapJSON:  tilemapJSON,
		tilesets:     tilesets,
		tilemapImage: tilemapImage,
		camera:       NewCamera(0.0, 0.0),
		colliders: []image.Rectangle{
			image.Rect(128, 128, 144, 144),
		},
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
