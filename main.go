package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1200
	screenHeight = 800
	worldWidth   = 2400
	worldHeight  = 1600
)

type Vector2 struct {
	X, Y float32
}

type GameObject struct {
	Position Vector2
	Size     float32
	Color    rl.Color
	Type     string
	Value    int
	Active   bool
	Rotation float32
}

type Hole struct {
	Position  Vector2
	Size      float32
	Score     int
	Speed     float32
	Animation float32
}

type Particle struct {
	Position Vector2
	Velocity Vector2
	Life     float32
	MaxLife  float32
	Color    rl.Color
	Size     float32
}

type Game struct {
	Player      Hole
	Objects     []GameObject
	Particles   []Particle
	Camera      rl.Camera2D
	GameTime    float32
	MaxGameTime float32
	BaseZoom    float32
}

func NewGame() *Game {
	game := &Game{
		Player: Hole{
			Position:  Vector2{X: worldWidth / 2, Y: worldHeight / 2},
			Size:      20.0,
			Score:     0,
			Speed:     200.0,
			Animation: 0.0,
		},
		Camera: rl.Camera2D{
			Offset:   rl.Vector2{X: screenWidth / 2, Y: screenHeight / 2},
			Target:   rl.Vector2{X: worldWidth / 2, Y: worldHeight / 2},
			Rotation: 0.0,
			Zoom:     1.0,
		},
		GameTime:    0.0,
		MaxGameTime: 120.0, // 2 minutes like the original
		BaseZoom:    1.0,
	}

	game.generateObjects()
	return game
}

func (g *Game) generateObjects() {
	rand.Seed(time.Now().UnixNano())

	// Generate small objects (people, benches, etc.)
	for i := 0; i < 200; i++ {
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     float32(2 + rand.Intn(4)),
			Color:    rl.Color{R: 139, G: 69, B: 19, A: 255}, // Saddle brown
			Type:     "small",
			Value:    1,
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate medium objects (cars, trees, etc.)
	for i := 0; i < 100; i++ {
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     float32(8 + rand.Intn(8)),
			Color:    rl.Color{R: 34, G: 139, B: 34, A: 255}, // Forest green
			Type:     "medium",
			Value:    5,
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate large objects (buildings, etc.)
	for i := 0; i < 50; i++ {
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     float32(20 + rand.Intn(30)),
			Color:    rl.Color{R: 105, G: 105, B: 105, A: 255}, // Dim gray
			Type:     "large",
			Value:    20,
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}
}

func (g *Game) addParticle(pos Vector2, color rl.Color) {
	for i := 0; i < 3; i++ {
		particle := Particle{
			Position: pos,
			Velocity: Vector2{
				X: (rand.Float32() - 0.5) * 100,
				Y: (rand.Float32() - 0.5) * 100,
			},
			Life:    1.0,
			MaxLife: 1.0,
			Color:   color,
			Size:    2 + rand.Float32()*3,
		}
		g.Particles = append(g.Particles, particle)
	}
}

func (g *Game) update(deltaTime float32) {
	g.GameTime += deltaTime
	g.Player.Animation += deltaTime * 2.0

	// Handle input
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		g.Player.Position.Y -= g.Player.Speed * deltaTime
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		g.Player.Position.Y += g.Player.Speed * deltaTime
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		g.Player.Position.X -= g.Player.Speed * deltaTime
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		g.Player.Position.X += g.Player.Speed * deltaTime
	}

	// Handle mouse movement
	mousePos := rl.GetMousePosition()
	screenCenter := Vector2{X: screenWidth / 2, Y: screenHeight / 2}
	direction := Vector2{
		X: mousePos.X - screenCenter.X,
		Y: mousePos.Y - screenCenter.Y,
	}

	// Normalize direction
	length := float32(math.Sqrt(float64(direction.X*direction.X + direction.Y*direction.Y)))
	if length > 0 {
		direction.X /= length
		direction.Y /= length

		// Move player towards mouse
		g.Player.Position.X += direction.X * g.Player.Speed * deltaTime
		g.Player.Position.Y += direction.Y * g.Player.Speed * deltaTime
	}

	// Keep player in bounds
	if g.Player.Position.X < g.Player.Size {
		g.Player.Position.X = g.Player.Size
	}
	if g.Player.Position.X > worldWidth-g.Player.Size {
		g.Player.Position.X = worldWidth - g.Player.Size
	}
	if g.Player.Position.Y < g.Player.Size {
		g.Player.Position.Y = g.Player.Size
	}
	if g.Player.Position.Y > worldHeight-g.Player.Size {
		g.Player.Position.Y = worldHeight - g.Player.Size
	}

	// Adaptive camera zoom based on hole size
	targetZoom := g.BaseZoom
	if g.Player.Size > 50 {
		// Gradually zoom out as hole gets bigger
		zoomFactor := 50.0 / g.Player.Size
		if zoomFactor < 0.2 {
			zoomFactor = 0.2 // Minimum zoom
		}
		targetZoom = zoomFactor
	}

	// Smooth zoom transition
	g.Camera.Zoom += (targetZoom - g.Camera.Zoom) * deltaTime * 2.0

	// Update camera to follow player
	g.Camera.Target = rl.Vector2{X: g.Player.Position.X, Y: g.Player.Position.Y}

	// Update particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		g.Particles[i].Life -= deltaTime
		g.Particles[i].Position.X += g.Particles[i].Velocity.X * deltaTime
		g.Particles[i].Position.Y += g.Particles[i].Velocity.Y * deltaTime
		g.Particles[i].Velocity.X *= 0.98 // Damping
		g.Particles[i].Velocity.Y *= 0.98

		if g.Particles[i].Life <= 0 {
			// Remove dead particle
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}

	// Animate object rotation
	for i := range g.Objects {
		if g.Objects[i].Active {
			g.Objects[i].Rotation += deltaTime * 30.0
		}
	}

	// Check collisions and consume objects
	for i := range g.Objects {
		if !g.Objects[i].Active {
			continue
		}

		// Calculate distance between hole and object
		dx := g.Player.Position.X - g.Objects[i].Position.X
		dy := g.Player.Position.Y - g.Objects[i].Position.Y
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		// Check if object can be consumed
		if distance < g.Player.Size && g.Player.Size > g.Objects[i].Size*0.8 {
			// Add particles at consumption point
			g.addParticle(g.Objects[i].Position, g.Objects[i].Color)

			g.Objects[i].Active = false
			g.Player.Score += g.Objects[i].Value

			// Grow the hole
			g.Player.Size += float32(g.Objects[i].Value) * 0.5
		}
	}
}

func (g *Game) drawGradientCircle(x, y, radius float32, innerColor, outerColor rl.Color) {
	steps := int32(radius / 2)
	if steps < 8 {
		steps = 8
	}
	if steps > 32 {
		steps = 32
	}

	for i := int32(0); i < steps; i++ {
		progress := float32(i) / float32(steps)
		currentRadius := radius * (1.0 - progress)

		// Interpolate color
		color := rl.Color{
			R: uint8(float32(innerColor.R)*(1.0-progress) + float32(outerColor.R)*progress),
			G: uint8(float32(innerColor.G)*(1.0-progress) + float32(outerColor.G)*progress),
			B: uint8(float32(innerColor.B)*(1.0-progress) + float32(outerColor.B)*progress),
			A: uint8(float32(innerColor.A)*(1.0-progress) + float32(outerColor.A)*progress),
		}

		rl.DrawCircle(int32(x), int32(y), currentRadius, color)
	}
}

func (g *Game) draw() {
	rl.BeginDrawing()

	// Gradient background
	rl.DrawRectangleGradientV(0, 0, screenWidth, screenHeight,
		rl.Color{R: 135, G: 206, B: 235, A: 255}, // Sky blue
		rl.Color{R: 25, G: 25, B: 112, A: 255})   // Midnight blue

	rl.BeginMode2D(g.Camera)

	// Draw world bounds with thicker, more visible border
	rl.DrawRectangleLinesEx(rl.Rectangle{X: 0, Y: 0, Width: worldWidth, Height: worldHeight}, 4, rl.White)

	// Draw objects with improved visuals
	for _, obj := range g.Objects {
		if obj.Active {
			// Draw shadow
			rl.DrawCircle(int32(obj.Position.X+2), int32(obj.Position.Y+2), obj.Size,
				rl.Color{R: 0, G: 0, B: 0, A: 50})

			// Draw main object with slight rotation effect
			if obj.Type == "small" {
				// People - draw as small rectangles
				rl.DrawRectanglePro(
					rl.Rectangle{X: obj.Position.X, Y: obj.Position.Y, Width: obj.Size, Height: obj.Size*1.5},
					rl.Vector2{X: obj.Size/2, Y: obj.Size*0.75},
					obj.Rotation,
					obj.Color)
			} else {
				// Trees and buildings - draw as circles with highlights
				rl.DrawCircle(int32(obj.Position.X), int32(obj.Position.Y), obj.Size, obj.Color)
				// Highlight
				rl.DrawCircle(int32(obj.Position.X-obj.Size*0.3), int32(obj.Position.Y-obj.Size*0.3),
					obj.Size*0.3, rl.Color{R: 255, G: 255, B: 255, A: 100})
			}
		}
	}

	// Draw particles
	for _, particle := range g.Particles {
		alpha := uint8(255.0 * (particle.Life / particle.MaxLife))
		color := particle.Color
		color.A = alpha
		rl.DrawCircle(int32(particle.Position.X), int32(particle.Position.Y), particle.Size, color)
	}

	// Draw player hole with enhanced visuals
	// Event horizon effect
	eventHorizon := g.Player.Size * 1.2
	g.drawGradientCircle(g.Player.Position.X, g.Player.Position.Y, eventHorizon,
		rl.Color{R: 0, G: 0, B: 0, A: 0},
		rl.Color{R: 50, G: 50, B: 50, A: 150})

	// Main black hole with pulsing effect
	pulse := 1.0 + float32(math.Sin(float64(g.Player.Animation)*3.0))*0.1
	g.drawGradientCircle(g.Player.Position.X, g.Player.Position.Y, g.Player.Size*pulse,
		rl.Color{R: 0, G: 0, B: 0, A: 255},
		rl.Color{R: 20, G: 20, B: 20, A: 255})

	// Inner core with swirling effect
	coreSize := g.Player.Size * 0.3
	swirl := g.Player.Animation * 100.0
	for i := 0; i < 8; i++ {
		angle := float64(i)*math.Pi/4.0 + float64(swirl)*0.01
		x := g.Player.Position.X + float32(math.Cos(angle))*coreSize*0.5
		y := g.Player.Position.Y + float32(math.Sin(angle))*coreSize*0.5
		rl.DrawCircle(int32(x), int32(y), 2, rl.Color{R: 100, G: 100, B: 100, A: 150})
	}

	rl.EndMode2D()

	// Enhanced UI
	uiColor := rl.White
	shadowColor := rl.Color{R: 0, G: 0, B: 0, A: 150}

	// Draw text shadows
	rl.DrawText(fmt.Sprintf("Score: %d", g.Player.Score), 12, 12, 24, shadowColor)
	rl.DrawText(fmt.Sprintf("Score: %d", g.Player.Score), 10, 10, 24, uiColor)

	rl.DrawText(fmt.Sprintf("Size: %.1f", g.Player.Size), 12, 42, 20, shadowColor)
	rl.DrawText(fmt.Sprintf("Size: %.1f", g.Player.Size), 10, 40, 20, uiColor)

	timeLeft := g.MaxGameTime - g.GameTime
	if timeLeft > 0 {
		timeColor := uiColor
		if timeLeft < 30 {
			// Flash red when time is running out
			flash := float32(math.Sin(float64(g.GameTime)*10.0))
			if flash > 0 {
				timeColor = rl.Red
			}
		}
		rl.DrawText(fmt.Sprintf("Time: %.1fs", timeLeft), 12, 72, 20, shadowColor)
		rl.DrawText(fmt.Sprintf("Time: %.1fs", timeLeft), 10, 70, 20, timeColor)
	} else {
		// Game over screen
		rl.DrawRectangle(0, 0, screenWidth, screenHeight, rl.Color{R: 0, G: 0, B: 0, A: 150})
		rl.DrawText("GAME OVER!", screenWidth/2-100, screenHeight/2-20, 40, rl.Red)
		rl.DrawText(fmt.Sprintf("Final Score: %d", g.Player.Score), screenWidth/2-80, screenHeight/2+30, 20, rl.White)
	}

	// Zoom indicator
	if g.Camera.Zoom < 0.8 {
		rl.DrawText(fmt.Sprintf("Zoom: %.1fx", g.Camera.Zoom), screenWidth-100, 10, 16, uiColor)
	}

	rl.DrawText("WASD or Mouse to move", 12, screenHeight-23, 16, shadowColor)
	rl.DrawText("WASD or Mouse to move", 10, screenHeight-25, 16, rl.Color{R: 200, G: 200, B: 200, A: 255})

	rl.EndDrawing()
}

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Hole.io Clone - Raylib Go")
	rl.SetTargetFPS(60)

	game := NewGame()

	for !rl.WindowShouldClose() {
		deltaTime := rl.GetFrameTime()

		if game.GameTime < game.MaxGameTime {
			game.update(deltaTime)
		}

		game.draw()
	}

	rl.CloseWindow()
}