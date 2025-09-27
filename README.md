# Hole.io Clone - Build Guide

A simple recreation of the popular hole.io game using raylib-go.

## What is Hole.io?

Hole.io is a multiplayer online game where players control a black hole and devour everything in their path to become the largest hole in the city. The core mechanics include:

- **Growing System**: Start small and consume objects to grow larger
- **Object Hierarchy**: Begin with small objects (people, benches) and progress to larger ones (cars, buildings)
- **Time Limit**: 2-minute rounds where the goal is to become the biggest hole
- **Physics-Based**: Objects fall into your hole based on size and proximity

## Game Features

This clone implements:

- ✅ Player-controlled black hole
- ✅ Object consumption mechanics
- ✅ Size-based growth system
- ✅ 2-minute game timer
- ✅ Score tracking
- ✅ Mouse and keyboard controls
- ✅ Camera following
- ✅ World boundaries
- ✅ Multiple object types with different values

## Prerequisites

Before building, ensure you have:

1. **Go 1.21+** installed
2. **C compiler** (gcc/clang) for raylib
3. **Platform-specific dependencies**:

### macOS
```bash
# Install Xcode command line tools
xcode-select --install

# Or install via Homebrew
brew install raylib
```

### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install build-essential git
sudo apt install libasound2-dev libgl1-mesa-dev libglu1-mesa-dev libxrandr-dev libxcursor-dev libxi-dev libxinerama-dev
```

### Windows
- Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MinGW-w64](https://www.mingw-w64.org/)
- Or use [Visual Studio Build Tools](https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022)

## Building the Game

1. **Clone/Download the project**:
   ```bash
   cd hole
   ```

2. **Initialize Go modules**:
   ```bash
   go mod tidy
   ```

3. **Build the game**:
   ```bash
   go build -o hole main.go
   ```

4. **Run the game**:
   ```bash
   ./hole
   ```

   On Windows:
   ```cmd
   hole.exe
   ```

## Alternative Build Methods

### Using Go Run
```bash
go run main.go
```

### Cross-compilation
```bash
# For Windows from Linux/macOS
GOOS=windows GOARCH=amd64 go build -o hole.exe main.go

# For Linux from macOS/Windows
GOOS=linux GOARCH=amd64 go build -o hole main.go
```

## Controls

- **WASD** or **Arrow Keys**: Move the hole
- **Mouse**: Move the hole toward cursor position
- **ESC**: Close the game

## Gameplay

1. **Start Small**: Begin as a tiny black hole
2. **Consume Objects**: Move over objects smaller than your hole
3. **Grow**: Each consumed object increases your size and score
4. **Progress**: Start with brown objects (people), then green (trees/cars), then gray (buildings)
5. **Time Limit**: You have 2 minutes to grow as large as possible

## Troubleshooting

### Common Issues

**Build errors with raylib**:
- Ensure you have a C compiler installed
- On Linux, install the required development libraries listed above
- Try updating Go: `go version` should be 1.21+

**"cannot find package" errors**:
```bash
go mod tidy
go clean -modcache
go mod download
```

**Performance issues**:
- The game targets 60 FPS
- Reduce object count in `generateObjects()` function if needed
- Adjust `screenWidth` and `screenHeight` constants for lower resolution

**Window doesn't appear**:
- Check if you have proper graphics drivers
- Try running from terminal to see error messages

### Platform-Specific Notes

**macOS**: You might need to allow the app in System Preferences > Security & Privacy

**Linux**: If using Wayland, you might need to set `export SDL_VIDEODRIVER=wayland`

**Windows**: Ensure all DLL dependencies are in the same folder as the executable

## Code Structure

- `main.go`: Complete game implementation
- `go.mod`: Go module definition with raylib-go dependency

## Future Enhancements

Possible improvements:
- Multiplayer support
- Different game modes
- Sound effects
- Better graphics and animations
- AI opponents
- Leaderboards

## Dependencies

- [raylib-go](https://github.com/gen2brain/raylib-go): Go bindings for raylib
- Go 1.21+ standard library