package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	screenWidth  = int32(1200)
	screenHeight = int32(800)
)

const (
	worldWidth  = 2400
	worldHeight = 1600
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

type NetworkPlayer struct {
	ID       int
	Hole     Hole
	Name     string
	Color    rl.Color
	LastSeen time.Time
}

type GameState int

const (
	StateMenu GameState = iota
	StateSinglePlayer
	StateMultiplayerHost
	StateMultiplayerClient
	StateLobby
	StateGameplay
	StateGameOver
)

type NetworkMessage struct {
	Type     string      `json:"type"`
	PlayerID int         `json:"player_id"`
	Data     interface{} `json:"data"`
}

type LobbyUpdate struct {
	PlayerCount int    `json:"player_count"`
	GameStarted bool   `json:"game_started"`
	HostReady   bool   `json:"host_ready"`
	ServerIP    string `json:"server_ip,omitempty"`
}

type PlayerUpdate struct {
	Position  Vector2 `json:"position"`
	Size      float32 `json:"size"`
	Score     int     `json:"score"`
	Animation float32 `json:"animation"`
}

type Game struct {
	State           GameState
	Player          Hole
	NetworkPlayers  map[int]*NetworkPlayer
	Objects         []GameObject
	Particles       []Particle
	Camera          rl.Camera2D
	GameTime        float32
	MaxGameTime     float32
	BaseZoom        float32
	MenuSelection   int
	IsHost          bool
	ServerConn      net.Conn
	ClientConns     []net.Conn
	PlayerID        int
	ServerIP        string
	InputText       string
	InputActive     bool
	LobbyReady      bool
	MinPlayers      int
	LocalIP         string
	GameStarted     bool
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipStr := ipNet.IP.String()
				// Prefer 192.168.x.x or 10.x.x.x ranges for LAN
				if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
					return ipStr
				}
			}
		}
	}
	return "127.0.0.1"
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	localIP := getLocalIP()
	return &Game{
		State:          StateMenu,
		NetworkPlayers: make(map[int]*NetworkPlayer),
		MenuSelection:  0,
		PlayerID:       rand.Intn(10000),
		ServerIP:       localIP + ":8080",
		LocalIP:        localIP,
		MinPlayers:     2,
		LobbyReady:     false,
		GameStarted:    false,
	}
}

func (g *Game) initSinglePlayer() {
	g.State = StateSinglePlayer
	g.Player = Hole{
		Position:  Vector2{X: worldWidth / 2, Y: worldHeight / 2},
		Size:      20.0,
		Score:     0,
		Speed:     200.0,
		Animation: 0.0,
	}
	g.Camera = rl.Camera2D{
		Offset:   rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2},
		Target:   rl.Vector2{X: worldWidth / 2, Y: worldHeight / 2},
		Rotation: 0.0,
		Zoom:     1.0,
	}
	g.GameTime = 0.0
	g.MaxGameTime = 120.0 // 2 minutes like the original
	g.BaseZoom = 1.0

	// Lock mouse cursor to the game window during gameplay
	rl.DisableCursor()

	g.generateObjects()
}

func (g *Game) generateObjects() {
	rand.Seed(time.Now().UnixNano())

	// Generate tiny objects (crumbs, coins, etc.) - easiest to eat
	for i := 0; i < 150; i++ {
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     float32(1 + rand.Intn(2)), // 1-2 size
			Color:    rl.Color{R: 255, G: 215, B: 0, A: 255}, // Gold
			Type:     "tiny",
			Value:    1,
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate small objects (people, pets, etc.)
	for i := 0; i < 200; i++ {
		size := float32(3 + rand.Intn(4)) // 3-6 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 139, G: 69, B: 19, A: 255}, // Saddle brown
			Type:     "small",
			Value:    int(size), // Value based on size
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate medium-small objects (bikes, benches, etc.)
	for i := 0; i < 120; i++ {
		size := float32(7 + rand.Intn(6)) // 7-12 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 0, G: 100, B: 0, A: 255}, // Dark green
			Type:     "medium-small",
			Value:    int(size),
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate medium objects (cars, small trees, etc.)
	for i := 0; i < 80; i++ {
		size := float32(13 + rand.Intn(8)) // 13-20 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 34, G: 139, B: 34, A: 255}, // Forest green
			Type:     "medium",
			Value:    int(size),
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate medium-large objects (trucks, large trees, etc.)
	for i := 0; i < 60; i++ {
		size := float32(21 + rand.Intn(12)) // 21-32 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 70, G: 130, B: 180, A: 255}, // Steel blue
			Type:     "medium-large",
			Value:    int(size),
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate large objects (small buildings, etc.)
	for i := 0; i < 40; i++ {
		size := float32(33 + rand.Intn(15)) // 33-47 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 105, G: 105, B: 105, A: 255}, // Dim gray
			Type:     "large",
			Value:    int(size),
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate extra large objects (medium buildings, etc.)
	for i := 0; i < 25; i++ {
		size := float32(48 + rand.Intn(20)) // 48-67 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 128, G: 128, B: 128, A: 255}, // Gray
			Type:     "extra-large",
			Value:    int(size),
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate huge objects (large buildings, etc.)
	for i := 0; i < 15; i++ {
		size := float32(68 + rand.Intn(25)) // 68-92 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 169, G: 169, B: 169, A: 255}, // Dark gray
			Type:     "huge",
			Value:    int(size),
			Active:   true,
			Rotation: rand.Float32() * 360,
		}
		g.Objects = append(g.Objects, obj)
	}

	// Generate massive objects (skyscrapers, etc.) - end game content
	for i := 0; i < 8; i++ {
		size := float32(93 + rand.Intn(30)) // 93-122 size
		obj := GameObject{
			Position: Vector2{
				X: rand.Float32() * worldWidth,
				Y: rand.Float32() * worldHeight,
			},
			Size:     size,
			Color:    rl.Color{R: 47, G: 79, B: 79, A: 255}, // Dark slate gray
			Type:     "massive",
			Value:    int(size),
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

func (g *Game) handleMenuInput() {
	if rl.IsKeyPressed(rl.KeyUp) {
		g.MenuSelection--
		if g.MenuSelection < 0 {
			g.MenuSelection = 2
		}
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		g.MenuSelection++
		if g.MenuSelection > 2 {
			g.MenuSelection = 0
		}
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		switch g.MenuSelection {
		case 0: // Single Player
			g.initSinglePlayer()
			g.State = StateGameplay
		case 1: // Host Multiplayer
			g.startServer()
			g.initSinglePlayer()
			g.State = StateLobby
		case 2: // Join Multiplayer
			g.InputActive = true
			g.InputText = g.ServerIP
		}
	}
}

func (g *Game) handleLobbyInput() {
	if rl.IsKeyPressed(rl.KeySpace) {
		g.LobbyReady = !g.LobbyReady
		if g.IsHost {
			// Host can start game if minimum players reached
			if len(g.NetworkPlayers)+1 >= g.MinPlayers && g.LobbyReady {
				g.startGame()
			}
		}
		g.sendLobbyUpdate()
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		// Return to menu
		g.State = StateMenu
		g.LobbyReady = false
		g.GameStarted = false
		// Release mouse cursor when returning to menu
		rl.EnableCursor()
		if g.ServerConn != nil {
			g.ServerConn.Close()
			g.ServerConn = nil
		}
	}
}

func (g *Game) sendLobbyUpdate() {
	update := LobbyUpdate{
		PlayerCount: len(g.NetworkPlayers) + 1,
		GameStarted: g.GameStarted,
		HostReady:   g.LobbyReady,
	}
	if g.IsHost {
		update.ServerIP = g.LocalIP + ":8080"
	}

	msg := NetworkMessage{
		Type:     "lobby_update",
		PlayerID: g.PlayerID,
		Data:     update,
	}

	data, _ := json.Marshal(msg)
	if g.IsHost {
		// Send to all clients
		for _, conn := range g.ClientConns {
			conn.Write(data)
			conn.Write([]byte("\n"))
		}
	} else if g.ServerConn != nil {
		// Send to server
		g.ServerConn.Write(data)
		g.ServerConn.Write([]byte("\n"))
	}
}

func (g *Game) startGame() {
	g.GameStarted = true
	g.State = StateGameplay
	g.GameTime = 0

	// Lock mouse cursor to the game window during multiplayer gameplay
	rl.DisableCursor()

	g.sendLobbyUpdate()
}

func (g *Game) handleTextInput() {
	key := rl.GetCharPressed()
	for key > 0 {
		if key >= 32 && key <= 125 && len(g.InputText) < 20 {
			g.InputText += string(rune(key))
		}
		key = rl.GetCharPressed()
	}
	if rl.IsKeyPressed(rl.KeyBackspace) && len(g.InputText) > 0 {
		g.InputText = g.InputText[:len(g.InputText)-1]
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		g.ServerIP = g.InputText
		g.connectToServer()
		g.InputActive = false
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		g.InputActive = false
	}
}

func (g *Game) startServer() {
	go func() {
		listener, err := net.Listen("tcp", ":8080")
		if err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
			return
		}
		defer listener.Close()
		fmt.Println("Server started on :8080")
		g.IsHost = true

		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			g.ClientConns = append(g.ClientConns, conn)
			go g.handleClient(conn)
		}
	}()
}

func (g *Game) connectToServer() {
	go func() {
		conn, err := net.Dial("tcp", g.ServerIP)
		if err != nil {
			fmt.Printf("Failed to connect to server: %v\n", err)
			return
		}
		g.ServerConn = conn
		g.initSinglePlayer()
		g.State = StateLobby
		go g.handleServerMessages()
		// Send initial lobby update to announce joining
		time.Sleep(100 * time.Millisecond) // Brief delay to ensure connection
		g.sendLobbyUpdate()
	}()
}

func (g *Game) handleClient(conn net.Conn) {
	// Send initial lobby state to new client
	g.sendLobbyUpdate()

	decoder := json.NewDecoder(conn)
	for {
		var msg NetworkMessage
		if err := decoder.Decode(&msg); err != nil {
			break
		}
		g.processNetworkMessage(msg)
		// Broadcast lobby updates to all clients when someone joins
		if msg.Type == "lobby_update" {
			g.sendLobbyUpdate()
		}
	}
	conn.Close()
}

func (g *Game) handleServerMessages() {
	decoder := json.NewDecoder(g.ServerConn)
	for {
		var msg NetworkMessage
		if err := decoder.Decode(&msg); err != nil {
			break
		}
		g.processNetworkMessage(msg)
	}
}

func (g *Game) processNetworkMessage(msg NetworkMessage) {
	switch msg.Type {
	case "player_update":
		data, _ := json.Marshal(msg.Data)
		var update PlayerUpdate
		json.Unmarshal(data, &update)
		if g.NetworkPlayers[msg.PlayerID] == nil {
			colors := []rl.Color{rl.Red, rl.Blue, rl.Green, rl.Yellow, rl.Purple, rl.Orange}
			g.NetworkPlayers[msg.PlayerID] = &NetworkPlayer{
				ID:    msg.PlayerID,
				Name:  fmt.Sprintf("Player %d", msg.PlayerID),
				Color: colors[msg.PlayerID%len(colors)],
			}
		}
		player := g.NetworkPlayers[msg.PlayerID]
		player.Hole.Position = update.Position
		player.Hole.Size = update.Size
		player.Hole.Score = update.Score
		player.Hole.Animation = update.Animation
		player.LastSeen = time.Now()
	case "lobby_update":
		data, _ := json.Marshal(msg.Data)
		var update LobbyUpdate
		json.Unmarshal(data, &update)
		// Add player to lobby if not already present
		if g.NetworkPlayers[msg.PlayerID] == nil {
			colors := []rl.Color{rl.Red, rl.Blue, rl.Green, rl.Yellow, rl.Purple, rl.Orange}
			g.NetworkPlayers[msg.PlayerID] = &NetworkPlayer{
				ID:    msg.PlayerID,
				Name:  fmt.Sprintf("Player %d", msg.PlayerID),
				Color: colors[msg.PlayerID%len(colors)],
				LastSeen: time.Now(),
			}
		} else {
			g.NetworkPlayers[msg.PlayerID].LastSeen = time.Now()
		}
		// If game started, transition to gameplay
		if update.GameStarted && g.State == StateLobby {
			g.State = StateGameplay
			g.GameTime = 0
		}
	}
}

func (g *Game) sendPlayerUpdate() {
	update := PlayerUpdate{
		Position:  g.Player.Position,
		Size:      g.Player.Size,
		Score:     g.Player.Score,
		Animation: g.Player.Animation,
	}
	msg := NetworkMessage{
		Type:     "player_update",
		PlayerID: g.PlayerID,
		Data:     update,
	}

	data, _ := json.Marshal(msg)
	if g.IsHost {
		// Send to all clients
		for _, conn := range g.ClientConns {
			conn.Write(data)
			conn.Write([]byte("\n"))
		}
	} else if g.ServerConn != nil {
		// Send to server
		g.ServerConn.Write(data)
		g.ServerConn.Write([]byte("\n"))
	}
}

func (g *Game) update(deltaTime float32) {
	switch g.State {
	case StateMenu:
		if g.InputActive {
			g.handleTextInput()
		} else {
			g.handleMenuInput()
		}
		return
	case StateLobby:
		g.handleLobbyInput()
		return
	case StateGameOver:
		g.handleGameOverInput()
		return
	case StateGameplay:
		// Continue with normal game update
		// Only update game time during gameplay
		if g.GameTime < g.MaxGameTime {
			g.GameTime += deltaTime
		}

		// Clean up old network players
		for id, player := range g.NetworkPlayers {
			if time.Since(player.LastSeen) > 5*time.Second {
				delete(g.NetworkPlayers, id)
			}
		}
		g.Player.Animation += deltaTime * 2.0

		// Check for game over and matchmaking
		if g.GameTime >= g.MaxGameTime {
			g.State = StateGameOver
			// Release mouse cursor when game ends
			rl.EnableCursor()
			return
		}
	default:
		return
	}

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
	screenCenter := Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2}
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

	// Update camera to follow player and handle window resizing
	g.Camera.Offset = rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2}
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

			// Grow the hole (heavily nerfed for longer progression)
			growthAmount := float32(g.Objects[i].Value) * 0.02 // Reduced from 0.5 to 0.02
			// Add diminishing returns for larger holes
			if g.Player.Size > 50 {
				growthAmount *= 0.7
			}
			if g.Player.Size > 100 {
				growthAmount *= 0.5
			}
			if g.Player.Size > 200 {
				growthAmount *= 0.3
			}
			g.Player.Size += growthAmount
		}
	}

	// Send network updates every 10 frames (6 times per second)
	if g.State == StateGameplay && (g.IsHost || g.ServerConn != nil) {
		if int(g.GameTime*60)%10 == 0 { // 60 FPS, every 10 frames
			g.sendPlayerUpdate()
		}
	}

}

func (g *Game) drawGradientCircle(x float32, y float32, radius float32, innerColor rl.Color, outerColor rl.Color) {
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

type PlayerResult struct {
	Name  string
	Size  float32
	Score int
}

func (g *Game) getGameResults() []PlayerResult {
	results := []PlayerResult{
		{Name: "You", Size: g.Player.Size, Score: g.Player.Score},
	}

	for _, player := range g.NetworkPlayers {
		results = append(results, PlayerResult{
			Name: player.Name,
			Size: player.Hole.Size,
			Score: player.Hole.Score,
		})
	}

	// Sort by size (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Size > results[i].Size {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

func (g *Game) handleGameOverInput() {
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		// If we were in multiplayer mode, return to lobby for easy LAN party mode
		if g.IsHost || g.ServerConn != nil {
			g.State = StateLobby
			// Reset game state but keep network connections
			g.GameTime = 0
			g.LobbyReady = false
			g.GameStarted = false
			// Generate new objects for next game
			g.generateObjects()
			// Reset player but keep network players connected
			g.Player = Hole{
				Position:  Vector2{X: worldWidth / 2, Y: worldHeight / 2},
				Size:      20.0,
				Score:     0,
				Speed:     200.0,
				Animation: 0.0,
			}
		} else {
			// Single player mode - return to menu
			g.State = StateMenu
			g.MenuSelection = 0
			// Reset for next match
			g.GameTime = 0
			g.NetworkPlayers = make(map[int]*NetworkPlayer)
			g.LobbyReady = false
			g.GameStarted = false
		}
	}
}

func (g *Game) drawMenu() {
	rl.BeginDrawing()

	// Gradient background
	rl.DrawRectangleGradientV(0, 0, screenWidth, screenHeight,
		rl.Color{R: 25, G: 25, B: 112, A: 255}, // Midnight blue
		rl.Color{R: 0, G: 0, B: 0, A: 255})     // Black

	// Title
	rl.DrawText("HOLE.IO CLONE", screenWidth/2-150, 100, 50, rl.White)
	rl.DrawText("Multiplayer Edition", screenWidth/2-120, 160, 25, rl.Gray)

	// Menu options
	menuOptions := []string{"Single Player", "Host Multiplayer", "Join Multiplayer"}
	for i, option := range menuOptions {
		y := 250 + i*60
		color := rl.White
		if i == g.MenuSelection {
			color = rl.Yellow
			rl.DrawText(">", screenWidth/2-200, int32(y), 30, rl.Yellow)
		}
		rl.DrawText(option, screenWidth/2-150, int32(y), 30, color)
	}

	// Input text box for IP address
	if g.InputActive {
		rl.DrawRectangle(screenWidth/2-150, 450, 300, 40, rl.Color{R: 50, G: 50, B: 50, A: 200})
		rl.DrawRectangleLines(screenWidth/2-150, 450, 300, 40, rl.White)
		rl.DrawText("Server IP:", screenWidth/2-140, 460, 20, rl.White)
		rl.DrawText(g.InputText, screenWidth/2-140, 480, 16, rl.LightGray)
		rl.DrawText("Press ENTER to connect, ESC to cancel", screenWidth/2-120, 500, 14, rl.Gray)
	}

	// Show LAN IP for hosting
	rl.DrawText(fmt.Sprintf("Your LAN IP: %s:8080", g.LocalIP), screenWidth/2-100, 550, 18, rl.Yellow)
	rl.DrawText("(Share this IP with friends to join your game)", screenWidth/2-140, 575, 14, rl.LightGray)

	// Instructions
	rl.DrawText("Use UP/DOWN arrows and ENTER to select", screenWidth/2-160, screenHeight-100, 18, rl.Gray)
	rl.DrawText("Timed matches - Top 3 players shown at end", screenWidth/2-170, screenHeight-70, 16, rl.DarkGray)
	rl.DrawText("LAN Multiplayer - Default port: 8080", screenWidth/2-140, screenHeight-40, 16, rl.DarkGray)

	rl.EndDrawing()
}

func (g *Game) drawLobby() {
	rl.BeginDrawing()

	// Gradient background
	rl.DrawRectangleGradientV(0, 0, screenWidth, screenHeight,
		rl.Color{R: 25, G: 25, B: 112, A: 255}, // Midnight blue
		rl.Color{R: 0, G: 0, B: 0, A: 255})     // Black

	// Title
	if g.IsHost {
		rl.DrawText("HOSTING LOBBY", screenWidth/2-120, 50, 40, rl.Yellow)
		rl.DrawText(fmt.Sprintf("Server: %s:8080", g.LocalIP), screenWidth/2-100, 100, 20, rl.White)
	} else {
		rl.DrawText("JOINED LOBBY", screenWidth/2-110, 50, 40, rl.Green)
		rl.DrawText(fmt.Sprintf("Connected to: %s", g.ServerIP), screenWidth/2-120, 100, 18, rl.White)
	}

	// Player list
	rl.DrawText("PLAYERS:", 50, 150, 30, rl.White)
	yPos := 190

	// Draw your player
	readyStatus := "NOT READY"
	readyColor := rl.Red
	if g.LobbyReady {
		readyStatus = "READY"
		readyColor = rl.Green
	}
	rl.DrawText(fmt.Sprintf("You (Player %d) - %s", g.PlayerID, readyStatus), 60, int32(yPos), 24, readyColor)
	yPos += 35

	// Draw network players
	for _, player := range g.NetworkPlayers {
		rl.DrawText(fmt.Sprintf("%s - CONNECTED", player.Name), 60, int32(yPos), 24, player.Color)
		yPos += 35
	}

	// Status and instructions
	playerCount := len(g.NetworkPlayers) + 1
	rl.DrawText(fmt.Sprintf("Players: %d/%d minimum", playerCount, g.MinPlayers), 50, 400, 20, rl.White)

	if g.IsHost {
		if playerCount >= g.MinPlayers {
			if g.LobbyReady {
				rl.DrawText("READY TO START! Game will begin shortly...", 50, 450, 20, rl.Green)
			} else {
				rl.DrawText("Press SPACE to ready up and start the game", 50, 450, 20, rl.Yellow)
			}
		} else {
			rl.DrawText(fmt.Sprintf("Waiting for %d more players...", g.MinPlayers-playerCount), 50, 450, 20, rl.Orange)
		}
	} else {
		if g.LobbyReady {
			rl.DrawText("READY - Waiting for host to start", 50, 450, 20, rl.Green)
		} else {
			rl.DrawText("Press SPACE to ready up", 50, 450, 20, rl.Yellow)
		}
	}

	// Controls
	rl.DrawText("SPACE - Ready/Unready", 50, screenHeight-80, 18, rl.Gray)
	rl.DrawText("ESC - Return to Menu", 50, screenHeight-50, 18, rl.Gray)

	// Connection indicator
	if g.IsHost {
		rl.DrawText("◆ HOST", screenWidth-100, 20, 20, rl.Yellow)
	} else {
		connStatus := "◆ CONNECTED"
		connColor := rl.Green
		if g.ServerConn == nil {
			connStatus = "◆ DISCONNECTED"
			connColor = rl.Red
		}
		rl.DrawText(connStatus, screenWidth-150, 20, 20, connColor)
	}

	rl.EndDrawing()
}

func (g *Game) drawGameOver() {
	rl.BeginDrawing()

	// Gradient background
	rl.DrawRectangleGradientV(0, 0, screenWidth, screenHeight,
		rl.Color{R: 25, G: 25, B: 112, A: 255}, // Midnight blue
		rl.Color{R: 0, G: 0, B: 0, A: 255})     // Black

	// Game Over title
	rl.DrawText("GAME OVER!", screenWidth/2-150, 50, 50, rl.Red)

	// Get results
	results := g.getGameResults()

	// Show final results
	rl.DrawText("FINAL RESULTS", screenWidth/2-120, 120, 30, rl.Yellow)

	// Show top 3 players prominently
	rl.DrawText("TOP 3 PLAYERS", screenWidth/2-100, 180, 25, rl.Yellow)

	yPos := 220
	for i := 0; i < 3 && i < len(results); i++ {
		result := results[i]

		rankColor := rl.White
		prefix := ""
		fontSize := int32(28)

		// Special styling for top 3
		switch i {
		case 0:
			rankColor = rl.Gold
			prefix = "🥇 WINNER! "
			fontSize = 32
		case 1:
			rankColor = rl.Color{R: 192, G: 192, B: 192, A: 255} // Silver
			prefix = "🥈 2nd Place "
		case 2:
			rankColor = rl.Color{R: 205, G: 127, B: 50, A: 255} // Bronze
			prefix = "🥉 3rd Place "
		}

		text := fmt.Sprintf("%s%s - Size: %.1f, Score: %d", prefix, result.Name, result.Size, result.Score)
		rl.DrawText(text, 50, int32(yPos), fontSize, rankColor)
		yPos += 50
	}

	// Show remaining players if any
	if len(results) > 3 {
		rl.DrawText("Other Players:", 50, int32(yPos+20), 20, rl.Gray)
		for i := 3; i < len(results) && i < 8; i++ {
			result := results[i]
			text := fmt.Sprintf("%d. %s - Size: %.1f, Score: %d", i+1, result.Name, result.Size, result.Score)
			rl.DrawText(text, 60, int32(yPos+50+(i-3)*25), 18, rl.LightGray)
		}
	}

	// Your final stats
	rl.DrawText("YOUR STATS:", 50, int32(yPos+40), 20, rl.Yellow)
	rl.DrawText(fmt.Sprintf("Final Size: %.1f", g.Player.Size), 60, int32(yPos+70), 18, rl.White)
	rl.DrawText(fmt.Sprintf("Final Score: %d", g.Player.Score), 60, int32(yPos+95), 18, rl.White)

	// Calculate rank
	rank := 1
	for _, result := range results {
		if result.Name != "You" && result.Size > g.Player.Size {
			rank++
		}
	}
	rl.DrawText(fmt.Sprintf("Your Rank: #%d", rank), 60, int32(yPos+120), 18, rl.Green)

	// Instructions
	rl.DrawText("Press ENTER or SPACE to return to menu", screenWidth/2-180, screenHeight-100, 20, rl.LightGray)

	rl.EndDrawing()
}

func (g *Game) draw() {
	if g.State == StateMenu {
		g.drawMenu()
		return
	}
	if g.State == StateLobby {
		g.drawLobby()
		return
	}
	if g.State == StateGameOver {
		g.drawGameOver()
		return
	}
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

			// Draw main object with type-specific rendering
			switch obj.Type {
			case "tiny":
				// Tiny objects - draw as small diamonds
				rl.DrawPoly(rl.Vector2{X: obj.Position.X, Y: obj.Position.Y}, 4, obj.Size, obj.Rotation, obj.Color)
			case "small":
				// People - draw as small rectangles
				rl.DrawRectanglePro(
					rl.Rectangle{X: obj.Position.X, Y: obj.Position.Y, Width: obj.Size, Height: obj.Size*1.5},
					rl.Vector2{X: obj.Size/2, Y: obj.Size*0.75},
					obj.Rotation,
					obj.Color)
			case "medium-small":
				// Bikes, benches - draw as hexagons
				rl.DrawPoly(rl.Vector2{X: obj.Position.X, Y: obj.Position.Y}, 6, obj.Size, obj.Rotation, obj.Color)
			default:
				// Medium and larger objects - draw as circles with highlights
				rl.DrawCircle(int32(obj.Position.X), int32(obj.Position.Y), obj.Size, obj.Color)
				// Highlight intensity based on size
				highlightAlpha := uint8(50 + (obj.Size * 2))
				if highlightAlpha > 150 {
					highlightAlpha = 150
				}
				rl.DrawCircle(int32(obj.Position.X-obj.Size*0.3), int32(obj.Position.Y-obj.Size*0.3),
					obj.Size*0.3, rl.Color{R: 255, G: 255, B: 255, A: highlightAlpha})
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

	// Draw network players
	for _, player := range g.NetworkPlayers {
		// Draw player hole with their color
		eventHorizon := player.Hole.Size * 1.2
		g.drawGradientCircle(player.Hole.Position.X, player.Hole.Position.Y, eventHorizon,
			rl.Color{R: 0, G: 0, B: 0, A: 0},
			rl.Color{R: player.Color.R / 4, G: player.Color.G / 4, B: player.Color.B / 4, A: 150})

		// Main hole with player color tint
		pulse := 1.0 + float32(math.Sin(float64(player.Hole.Animation)*3.0))*0.1
		g.drawGradientCircle(player.Hole.Position.X, player.Hole.Position.Y, player.Hole.Size*pulse,
			rl.Color{R: 0, G: 0, B: 0, A: 255},
			rl.Color{R: player.Color.R / 8, G: player.Color.G / 8, B: player.Color.B / 8, A: 255})

		// Player name tag
		nameX := player.Hole.Position.X - float32(len(player.Name)*3)
		nameY := player.Hole.Position.Y - player.Hole.Size - 20
		rl.DrawText(player.Name, int32(nameX), int32(nameY), 16, player.Color)
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

	// Show multiplayer info
	if len(g.NetworkPlayers) > 0 {
		rl.DrawText(fmt.Sprintf("Players: %d", len(g.NetworkPlayers)+1), screenWidth-120, 12, 18, shadowColor)
		rl.DrawText(fmt.Sprintf("Players: %d", len(g.NetworkPlayers)+1), screenWidth-122, 10, 18, uiColor)
	}

	rl.DrawText("WASD or Mouse to move", 12, screenHeight-23, 16, shadowColor)
	rl.DrawText("WASD or Mouse to move", 10, screenHeight-25, 16, rl.Color{R: 200, G: 200, B: 200, A: 255})

	rl.EndDrawing()
}

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Hole.io Clone - Raylib Go")
	rl.SetWindowState(rl.FlagWindowResizable)
	rl.SetTargetFPS(60)

	game := NewGame()

	for !rl.WindowShouldClose() {
		deltaTime := rl.GetFrameTime()

		// Update screen dimensions if window was resized
		if rl.IsWindowResized() {
			screenWidth = int32(rl.GetScreenWidth())
			screenHeight = int32(rl.GetScreenHeight())
		}

		// Always update, but handle different states inside update function
		game.update(deltaTime)

		game.draw()
	}

	rl.CloseWindow()
}