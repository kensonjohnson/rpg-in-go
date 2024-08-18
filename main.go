package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"rpg-tutorial/entities"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	player       *entities.Player
	enemies      []*entities.Enemy
	potions      []*entities.Potion
	tilemapJSON  *TilemapJSON
	tilemapImage *ebiten.Image
	camera       *Camera
}

func (g *Game) Update() error {
	// React to keypressed
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Y -= 2

	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Y += 2
	}

	for _, sprite := range g.enemies {
		if sprite.FollowsPlayer {
			if sprite.X < g.player.X {
				sprite.X += 1
			} else if sprite.X > g.player.X {
				sprite.X -= 1
			}
			if sprite.Y < g.player.Y {
				sprite.Y += 1
			} else if sprite.Y > g.player.Y {
				sprite.Y -= 1
			}
		}
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
	for _, layer := range g.tilemapJSON.Layers {
		for index, id := range layer.Data {
			x := index % layer.Width
			y := index / layer.Width

			x *= 16
			y *= 16

			srcX := (id - 1) % 22
			srcY := (id - 1) / 22

			srcX *= 16
			srcY *= 16

			options.GeoM.Translate(float64(x), float64(y))
			options.GeoM.Translate(g.camera.X, g.camera.Y)

			screen.DrawImage(
				g.tilemapImage.SubImage(image.Rect(srcX, srcY, srcX+16, srcY+16)).(*ebiten.Image),
				&options,
			)

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
		tilemapImage: tilemapImage,
		camera:       NewCamera(0.0, 0.0),
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
