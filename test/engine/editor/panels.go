// Panel implementations for the game engine editor
package editor

import (
	"fmt"

	"gameengine/components"
	"gameengine/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Panel interface for all editor panels
type Panel interface {
	Initialize() error
	Update(deltaTime float32)
	Render(rect rl.Rectangle)
	Shutdown()
}

// SceneHierarchyPanel displays the scene hierarchy
type SceneHierarchyPanel struct {
	editor          *Editor
	scrollOffset    rl.Vector2
	expandedNodes   map[string]bool
	searchText      string
	searchTextBuf   []byte
	entityNames     map[core.EntityID]string  // Cache entity names to prevent recalculation
	lastFrameCount  uint64                     // Track frame count to know when to update cache
}

// NewSceneHierarchyPanel creates a new scene hierarchy panel
func NewSceneHierarchyPanel(editor *Editor) *SceneHierarchyPanel {
	return &SceneHierarchyPanel{
		editor:        editor,
		expandedNodes: make(map[string]bool),
		searchTextBuf: make([]byte, 256),
		entityNames:   make(map[core.EntityID]string),
		lastFrameCount: 0,
	}
}

func (p *SceneHierarchyPanel) Initialize() error {
	return nil
}

func (p *SceneHierarchyPanel) Update(deltaTime float32) {
	// Update logic for hierarchy panel
}

func (p *SceneHierarchyPanel) Render(rect rl.Rectangle) {
	// Draw panel background
	rl.DrawRectangleRec(rect, rl.Color{R: 50, G: 50, B: 50, A: 255})
	rl.DrawRectangleLinesEx(rect, 1, rl.Color{R: 70, G: 70, B: 70, A: 255})

	// Panel title
	titleHeight := float32(25)
	titleRect := rl.Rectangle{X: rect.X, Y: rect.Y, Width: rect.Width, Height: titleHeight}
	rl.DrawRectangleRec(titleRect, rl.Color{R: 60, G: 60, B: 60, A: 255})
	rl.DrawText("Scene Hierarchy", int32(rect.X + 10), int32(rect.Y + 5), 12, rl.White)

	// Search bar
	searchHeight := float32(25)
	// searchRect := rl.Rectangle{X: rect.X + 5, Y: rect.Y + titleHeight + 5, Width: rect.Width - 10, Height: searchHeight}
	// Search text box (commented out - raygui disabled)
	// if rg.GuiTextBox(searchRect, p.searchTextBuf, 256, true) {
	//	p.searchText = string(p.searchTextBuf[:p.findNullTerminator(p.searchTextBuf)])
	// }

	// Entity list area
	listRect := rl.Rectangle{
		X: rect.X + 5,
		Y: rect.Y + titleHeight + searchHeight + 10,
		Width: rect.Width - 10,
		Height: rect.Height - titleHeight - searchHeight - 15,
	}

	p.renderEntityList(listRect)
}

func (p *SceneHierarchyPanel) renderEntityList(rect rl.Rectangle) {
	// Clear the list area background first to prevent flickering
	rl.DrawRectangleRec(rect, rl.Color{R: 50, G: 50, B: 50, A: 255})

	activeScene := p.editor.gameEngine.GetSceneManager().GetActiveScene()
	if activeScene == nil {
		rl.DrawText("No active scene", int32(rect.X + 10), int32(rect.Y + 10), 10, rl.Gray)
		return
	}

	// Get all entities (simplified - get all entities with any component)
	world := activeScene.GetWorld()
	entities := world.GetEntitiesWithComponent(components.TransformComponentType)

	itemHeight := float32(20)
	y := rect.Y

	// Static entity names to completely eliminate dynamic string operations
	staticNames := []string{"Entity 1", "Entity 2", "Entity 3"}

	for i, entityID := range entities {
		if y + itemHeight > rect.Y + rect.Height {
			break // Don't render beyond panel bounds
		}

		itemRect := rl.Rectangle{X: rect.X, Y: y, Width: rect.Width, Height: itemHeight}

		// Check if entity is selected
		isSelected := p.editor.selectedEntity == entityID

		// Draw stable background
		backgroundColor := rl.Color{R: 50, G: 50, B: 50, A: 255} // Default background
		if isSelected {
			backgroundColor = rl.Color{R: 0, G: 120, B: 215, A: 255} // Selection color
		}

		// Draw item background
		rl.DrawRectangleRec(itemRect, backgroundColor)

		// Handle click (check mouse position only when clicking)
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			mousePos := rl.GetMousePosition()
			if rl.CheckCollisionPointRec(mousePos, itemRect) {
				p.editor.SetSelectedEntity(entityID)
			}
		}

		// Use static names to eliminate ALL dynamic text operations
		var entityName string
		if i < len(staticNames) {
			entityName = staticNames[i]
		} else {
			entityName = "Entity N"
		}

		// Draw text with fixed positioning
		rl.DrawText(entityName, int32(rect.X + 10), int32(y + 2), 10, rl.White)

		y += itemHeight
	}
}

func (p *SceneHierarchyPanel) findNullTerminator(buf []byte) int {
	for i, b := range buf {
		if b == 0 {
			return i
		}
	}
	return len(buf)
}

func (p *SceneHierarchyPanel) Shutdown() {
	// Cleanup
}

// InspectorPanel displays properties of the selected entity
type InspectorPanel struct {
	editor        *Editor
	scrollOffset  rl.Vector2
	textBuffers   map[string][]byte
}

// NewInspectorPanel creates a new inspector panel
func NewInspectorPanel(editor *Editor) *InspectorPanel {
	return &InspectorPanel{
		editor:      editor,
		textBuffers: make(map[string][]byte),
	}
}

func (p *InspectorPanel) Initialize() error {
	return nil
}

func (p *InspectorPanel) Update(deltaTime float32) {
	// Update logic for inspector panel
}

func (p *InspectorPanel) Render(rect rl.Rectangle) {
	// Draw panel background
	rl.DrawRectangleRec(rect, rl.Color{R: 50, G: 50, B: 50, A: 255})
	rl.DrawRectangleLinesEx(rect, 1, rl.Color{R: 70, G: 70, B: 70, A: 255})

	// Panel title
	titleHeight := float32(25)
	titleRect := rl.Rectangle{X: rect.X, Y: rect.Y, Width: rect.Width, Height: titleHeight}
	rl.DrawRectangleRec(titleRect, rl.Color{R: 60, G: 60, B: 60, A: 255})
	rl.DrawText("Inspector", int32(rect.X + 10), int32(rect.Y + 5), 12, rl.White)

	// Content area
	contentRect := rl.Rectangle{
		X: rect.X + 5,
		Y: rect.Y + titleHeight + 5,
		Width: rect.Width - 10,
		Height: rect.Height - titleHeight - 10,
	}

	if p.editor.selectedEntity == 0 {
		rl.DrawText("No entity selected", int32(contentRect.X + 10), int32(contentRect.Y + 10), 10, rl.Gray)
		return
	}

	p.renderEntityInspector(contentRect)
}

func (p *InspectorPanel) renderEntityInspector(rect rl.Rectangle) {
	activeScene := p.editor.gameEngine.GetSceneManager().GetActiveScene()
	if activeScene == nil {
		return
	}

	world := activeScene.GetWorld()
	entityID := p.editor.selectedEntity

	// Entity header
	y := rect.Y
	headerHeight := float32(30)

	// Entity name and active checkbox
	entityName := fmt.Sprintf("Entity %d", entityID)
	rl.DrawText(entityName, int32(rect.X + 10), int32(y + 5), 14, rl.White)

	// Active checkbox (commented out - raygui disabled)
	// activeRect := rl.Rectangle{X: rect.X + rect.Width - 60, Y: y + 5, Width: 50, Height: 20}
	// entityActive := true // Would get from entity state
	// rg.GuiCheckBox(activeRect, "Active", &entityActive)

	y += headerHeight

	// Transform component (every entity should have one)
	if transform, ok := world.GetComponent(entityID, components.TransformComponentType); ok {
		y = p.renderTransformComponent(rect, y, transform.(*components.TransformComponent))
	}

	// Render other components (simplified - just show that there are components)
	// componentTypes := world.GetEntityComponentTypes(entityID)  // Method doesn't exist
	// for _, componentType := range componentTypes {
	if false { // Disabled for now
		// if componentType == components.TransformComponentType {
		//	continue // Already rendered
		// }

		// component, _ := world.GetComponent(entityID, componentType)
		// y = p.renderGenericComponent(rect, y, componentType, component)

		// if y > rect.Y + rect.Height {
		//	break // Don't render beyond panel bounds
		// }
	}
}

func (p *InspectorPanel) renderTransformComponent(rect rl.Rectangle, y float32, transform *components.TransformComponent) float32 {
	// sectionHeight := float32(120)  // Unused variable

	// Component header
	headerRect := rl.Rectangle{X: rect.X, Y: y, Width: rect.Width, Height: 25}
	rl.DrawRectangleRec(headerRect, rl.Color{R: 65, G: 65, B: 65, A: 255})
	rl.DrawText("Transform", int32(rect.X + 10), int32(y + 5), 12, rl.White)

	y += 30

	// Position
	rl.DrawText("Position", int32(rect.X + 10), int32(y), 10, rl.LightGray)
	y += 15

	pos := transform.Position
	p.renderVector3Input(rect, y, "position", &pos)
	transform.SetPosition(pos)
	y += 25

	// Rotation
	rl.DrawText("Rotation", int32(rect.X + 10), int32(y), 10, rl.LightGray)
	y += 15

	rot := transform.Rotation
	p.renderVector3Input(rect, y, "rotation", &rot)
	transform.SetRotation(rot)
	y += 25

	// Scale
	rl.DrawText("Scale", int32(rect.X + 10), int32(y), 10, rl.LightGray)
	y += 15

	scale := transform.Scale
	p.renderVector3Input(rect, y, "scale", &scale)
	transform.SetScale(scale)
	y += 25

	return y
}

func (p *InspectorPanel) renderVector3Input(rect rl.Rectangle, y float32, name string, vec *rl.Vector3) {
	fieldWidth := (rect.Width - 30) / 3
	spacing := float32(5)

	// X (text input disabled - raygui not available)
	// xBuf := p.getOrCreateTextBuffer(name+"_x", fmt.Sprintf("%.2f", vec.X))
	// xRect := rl.Rectangle{X: rect.X + 10, Y: y, Width: fieldWidth, Height: 20}

	// Draw current values as text instead
	rl.DrawText(fmt.Sprintf("X: %.2f", vec.X), int32(rect.X + 10), int32(y), 10, rl.White)
	// X input field (commented out - raygui disabled)
	// if rg.GuiTextBox(xRect, xBuf, 32, true) {
	//	if val, err := strconv.ParseFloat(strings.TrimSpace(string(xBuf[:p.findNullTerminator(xBuf)])), 32); err == nil {
	//		vec.X = float32(val)
	//	}
	// }

	// Y (text input disabled - raygui not available)
	// yBuf := p.getOrCreateTextBuffer(name+"_y", fmt.Sprintf("%.2f", vec.Y))
	// yRect := rl.Rectangle{X: rect.X + 10 + fieldWidth + spacing, Y: y, Width: fieldWidth, Height: 20}

	// Draw current values as text instead
	rl.DrawText(fmt.Sprintf("Y: %.2f", vec.Y), int32(rect.X + 10 + fieldWidth + spacing), int32(y), 10, rl.White)
	// Y input field (commented out - raygui disabled)
	// if rg.GuiTextBox(yRect, yBuf, 32, true) {
	//	if val, err := strconv.ParseFloat(strings.TrimSpace(string(yBuf[:p.findNullTerminator(yBuf)])), 32); err == nil {
	//		vec.Y = float32(val)
	//	}
	// }

	// Z (text input disabled - raygui not available)
	// zBuf := p.getOrCreateTextBuffer(name+"_z", fmt.Sprintf("%.2f", vec.Z))
	// zRect := rl.Rectangle{X: rect.X + 10 + 2*(fieldWidth + spacing), Y: y, Width: fieldWidth, Height: 20}

	// Draw current values as text instead
	rl.DrawText(fmt.Sprintf("Z: %.2f", vec.Z), int32(rect.X + 10 + 2*(fieldWidth + spacing)), int32(y), 10, rl.White)
	// Z input field (commented out - raygui disabled)
	// if rg.GuiTextBox(zRect, zBuf, 32, true) {
	//	if val, err := strconv.ParseFloat(strings.TrimSpace(string(zBuf[:p.findNullTerminator(zBuf)])), 32); err == nil {
	//		vec.Z = float32(val)
	//	}
	// }
}

func (p *InspectorPanel) renderGenericComponent(rect rl.Rectangle, y float32, componentType core.ComponentType, component interface{}) float32 {
	// Component header
	headerRect := rl.Rectangle{X: rect.X, Y: y, Width: rect.Width, Height: 25}
	rl.DrawRectangleRec(headerRect, rl.Color{R: 65, G: 65, B: 65, A: 255})

	componentName := fmt.Sprintf("Component %d", componentType)
	rl.DrawText(componentName, int32(rect.X + 10), int32(y + 5), 12, rl.White)

	y += 30

	// Basic component info (would be expanded based on component type)
	rl.DrawText("Component data...", int32(rect.X + 10), int32(y), 10, rl.Gray)
	y += 20

	return y
}

func (p *InspectorPanel) getOrCreateTextBuffer(key string, defaultValue string) []byte {
	if buf, exists := p.textBuffers[key]; exists {
		return buf
	}

	buf := make([]byte, 256)
	copy(buf, []byte(defaultValue))
	p.textBuffers[key] = buf
	return buf
}

func (p *InspectorPanel) findNullTerminator(buf []byte) int {
	for i, b := range buf {
		if b == 0 {
			return i
		}
	}
	return len(buf)
}

func (p *InspectorPanel) Shutdown() {
	// Cleanup
}

// ViewportPanel renders the 3D scene
type ViewportPanel struct {
	editor         *Editor
	renderTexture  rl.RenderTexture2D
	viewportSize   rl.Vector2
}

// NewViewportPanel creates a new viewport panel
func NewViewportPanel(editor *Editor) *ViewportPanel {
	return &ViewportPanel{
		editor: editor,
	}
}

func (p *ViewportPanel) Initialize() error {
	// Initialize render texture for viewport
	p.viewportSize = rl.Vector2{X: 800, Y: 600}
	p.renderTexture = rl.LoadRenderTexture(int32(p.viewportSize.X), int32(p.viewportSize.Y))
	return nil
}

func (p *ViewportPanel) Update(deltaTime float32) {
	// Update viewport logic
}

func (p *ViewportPanel) Render(rect rl.Rectangle) {
	// Draw panel background
	rl.DrawRectangleRec(rect, rl.Color{R: 50, G: 50, B: 50, A: 255})
	rl.DrawRectangleLinesEx(rect, 1, rl.Color{R: 70, G: 70, B: 70, A: 255})

	// Panel title
	titleHeight := float32(25)
	titleRect := rl.Rectangle{X: rect.X, Y: rect.Y, Width: rect.Width, Height: titleHeight}
	rl.DrawRectangleRec(titleRect, rl.Color{R: 60, G: 60, B: 60, A: 255})
	rl.DrawText("Scene View", int32(rect.X + 10), int32(rect.Y + 5), 12, rl.White)

	// Define viewport area
	viewportRect := rl.Rectangle{
		X: rect.X + 5,
		Y: rect.Y + titleHeight + 5,
		Width: rect.Width - 10,
		Height: rect.Height - titleHeight - 10,
	}

	// Draw viewport background
	rl.DrawRectangleRec(viewportRect, rl.Color{R: 100, G: 149, B: 237, A: 255}) // Sky blue

	// Set scissor test to clip 3D rendering to viewport area
	rl.BeginScissorMode(int32(viewportRect.X), int32(viewportRect.Y), int32(viewportRect.Width), int32(viewportRect.Height))

	// Begin 3D mode with editor camera
	rl.BeginMode3D(*p.editor.GetEditorCamera())

	// Render grid if enabled
	if p.editor.showGrid {
		p.renderGrid()
	}

	// Render scene entities
	p.renderSceneEntities()

	// Render gizmos if entity is selected
	if p.editor.selectedEntity != 0 {
		p.renderGizmos()
	}

	rl.EndMode3D()
	rl.EndScissorMode()
}

func (p *ViewportPanel) renderGrid() {
	gridSize := float32(p.editor.gridSize)
	spacing := p.editor.gridSpacing

	rl.DrawGrid(int32(gridSize), spacing)
}

func (p *ViewportPanel) renderSceneEntities() {
	activeScene := p.editor.gameEngine.GetSceneManager().GetActiveScene()
	if activeScene == nil {
		return
	}

	world := activeScene.GetWorld()

	// Get entities with mesh renderer components
	entities := world.GetEntitiesWithComponents(components.TransformComponentType, components.MeshRendererComponentType)

	for _, entityID := range entities {
		transform, _ := world.GetComponent(entityID, components.TransformComponentType)
		meshRenderer, _ := world.GetComponent(entityID, components.MeshRendererComponentType)

		if transform != nil && meshRenderer != nil {
			transformComp := transform.(*components.TransformComponent)
			// meshComp := meshRenderer.(*components.MeshRendererComponent)  // Unused

			// Set up transform matrix
			position := transformComp.Position
			// rotation := transformComp.Rotation  // Unused
			scale := transformComp.Scale

			// Draw the model (simplified - would use proper rendering system)
			// model := meshComp.GetModel()  // Method doesn't exist yet
			// rl.DrawModelEx(model, position, rl.Vector3{X: 0, Y: 1, Z: 0}, rotation.Y, scale, rl.White)
			// For now, just draw a cube placeholder
			rl.DrawCube(position, scale.X, scale.Y, scale.Z, rl.White)
		}
	}
}

func (p *ViewportPanel) renderGizmos() {
	activeScene := p.editor.gameEngine.GetSceneManager().GetActiveScene()
	if activeScene == nil {
		return
	}

	world := activeScene.GetWorld()

	if transform, ok := world.GetComponent(p.editor.selectedEntity, components.TransformComponentType); ok {
		transformComp := transform.(*components.TransformComponent)
		position := transformComp.Position

		// Simple gizmo rendering (basic axes)
		gizmoSize := float32(1.0)

		switch p.editor.gizmoMode {
		case GizmoModeTranslate:
			// X axis (red)
			rl.DrawLine3D(position, rl.Vector3Add(position, rl.Vector3{X: gizmoSize, Y: 0, Z: 0}), rl.Red)
			// Y axis (green)
			rl.DrawLine3D(position, rl.Vector3Add(position, rl.Vector3{X: 0, Y: gizmoSize, Z: 0}), rl.Green)
			// Z axis (blue)
			rl.DrawLine3D(position, rl.Vector3Add(position, rl.Vector3{X: 0, Y: 0, Z: gizmoSize}), rl.Blue)
		}
	}
}

func (p *ViewportPanel) Shutdown() {
	rl.UnloadRenderTexture(p.renderTexture)
}

// ProjectBrowserPanel shows project files and assets
type ProjectBrowserPanel struct {
	editor *Editor
}

// NewProjectBrowserPanel creates a new project browser panel
func NewProjectBrowserPanel(editor *Editor) *ProjectBrowserPanel {
	return &ProjectBrowserPanel{
		editor: editor,
	}
}

func (p *ProjectBrowserPanel) Initialize() error {
	return nil
}

func (p *ProjectBrowserPanel) Update(deltaTime float32) {
	// Update logic
}

func (p *ProjectBrowserPanel) Render(rect rl.Rectangle) {
	// Draw panel background
	rl.DrawRectangleRec(rect, rl.Color{R: 50, G: 50, B: 50, A: 255})
	rl.DrawRectangleLinesEx(rect, 1, rl.Color{R: 70, G: 70, B: 70, A: 255})

	// Panel title
	titleHeight := float32(25)
	titleRect := rl.Rectangle{X: rect.X, Y: rect.Y, Width: rect.Width, Height: titleHeight}
	rl.DrawRectangleRec(titleRect, rl.Color{R: 60, G: 60, B: 60, A: 255})
	rl.DrawText("Project", int32(rect.X + 10), int32(rect.Y + 5), 12, rl.White)

	// Content area
	rl.DrawText("Project browser coming soon...", int32(rect.X + 10), int32(rect.Y + titleHeight + 10), 10, rl.Gray)
}

func (p *ProjectBrowserPanel) Shutdown() {
	// Cleanup
}

// ConsolePanel shows debug console and logs
type ConsolePanel struct {
	editor *Editor
}

// NewConsolePanel creates a new console panel
func NewConsolePanel(editor *Editor) *ConsolePanel {
	return &ConsolePanel{
		editor: editor,
	}
}

func (p *ConsolePanel) Initialize() error {
	return nil
}

func (p *ConsolePanel) Update(deltaTime float32) {
	// Update logic
}

func (p *ConsolePanel) Render(rect rl.Rectangle) {
	// Draw panel background
	rl.DrawRectangleRec(rect, rl.Color{R: 50, G: 50, B: 50, A: 255})
	rl.DrawRectangleLinesEx(rect, 1, rl.Color{R: 70, G: 70, B: 70, A: 255})

	// Panel title
	titleHeight := float32(25)
	titleRect := rl.Rectangle{X: rect.X, Y: rect.Y, Width: rect.Width, Height: titleHeight}
	rl.DrawRectangleRec(titleRect, rl.Color{R: 60, G: 60, B: 60, A: 255})
	rl.DrawText("Console", int32(rect.X + 10), int32(rect.Y + 5), 12, rl.White)

	// Content area
	rl.DrawText("Console output will appear here...", int32(rect.X + 10), int32(rect.Y + titleHeight + 10), 10, rl.Gray)
}

func (p *ConsolePanel) Shutdown() {
	// Cleanup
}