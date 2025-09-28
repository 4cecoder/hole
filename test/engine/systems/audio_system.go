// Package systems provides audio system implementation
package systems

import (
	"fmt"
	"gameengine/components"
	"gameengine/core"
	"gameengine/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// AudioSystem handles 3D audio processing and playback
type AudioSystem struct {
	world           *ecs.World
	masterVolume    float32
	listenerEntity  core.EntityID
	maxAudioSources int
	activeAudioSources []ActiveAudioSource
	reverbZones     []ReverbZoneData
	initialized     bool
	audioDevice     bool
	sampleRate      int
	bufferSize      int
	channels        int
	distanceModel   DistanceModel
	dopplerEnabled  bool
}

// ActiveAudioSource tracks currently playing audio sources
type ActiveAudioSource struct {
	EntityID     core.EntityID
	AudioSource  *components.AudioSourceComponent
	Transform    *components.TransformComponent
	Priority     int
	Distance     float32
	Volume       float32
	IsAudible    bool
	LastPosition rl.Vector3
	Velocity     rl.Vector3
}

// ReverbZoneData contains reverb zone information
type ReverbZoneData struct {
	EntityID    core.EntityID
	ReverbZone  *components.AudioReverbZoneComponent
	Transform   *components.TransformComponent
	Influence   float32
}

// DistanceModel defines how audio volume changes with distance
type DistanceModel int

const (
	InverseDistance DistanceModel = iota
	InverseDistanceClamped
	LinearDistance
	LinearDistanceClamped
	ExponentDistance
	ExponentDistanceClamped
	NoDistanceAttenuation
)

// NewAudioSystem creates a new audio system
func NewAudioSystem(world *ecs.World) *AudioSystem {
	return &AudioSystem{
		world:              world,
		masterVolume:       1.0,
		listenerEntity:     0,
		maxAudioSources:    32,
		activeAudioSources: make([]ActiveAudioSource, 0, 32),
		reverbZones:        make([]ReverbZoneData, 0, 8),
		initialized:        false,
		audioDevice:        false,
		sampleRate:         44100,
		bufferSize:         1024,
		channels:           2,
		distanceModel:      InverseDistanceClamped,
		dopplerEnabled:     true,
	}
}

// Initialize initializes the audio system
func (as *AudioSystem) Initialize() error {
	// Check if audio device is already initialized to avoid double initialization
	if rl.IsAudioDeviceReady() {
		// Audio device is already ready, just mark as initialized
		as.audioDevice = true
		as.initialized = true
		return nil
	}

	// Initialize audio device only if not already ready
	rl.InitAudioDevice()

	// Verify initialization was successful
	if !rl.IsAudioDeviceReady() {
		as.audioDevice = false
		as.initialized = false
		return fmt.Errorf("failed to initialize audio device")
	}

	as.audioDevice = true
	as.initialized = true
	return nil
}

// Shutdown shuts down the audio system
func (as *AudioSystem) Shutdown() {
	if as.audioDevice {
		rl.CloseAudioDevice()
		as.audioDevice = false
	}
	as.initialized = false
}

// GetPriority returns the system priority
func (as *AudioSystem) GetPriority() core.SystemPriority {
	return core.PriorityAudio
}

// Update processes audio for all entities
func (as *AudioSystem) Update(deltaTime float32) {
	if !as.initialized || !as.audioDevice {
		return
	}

	// Find the audio listener
	as.findAudioListener()

	// Update audio sources
	as.updateAudioSources(deltaTime)

	// Update reverb zones
	as.updateReverbZones()

	// Process 3D audio
	as.process3DAudio(deltaTime)

	// Update sound playback
	as.updateSoundPlayback(deltaTime)
}

// findAudioListener finds the active audio listener
func (as *AudioSystem) findAudioListener() {
	listenerEntities := as.world.GetEntitiesWithComponents(components.AudioListenerComponentType, components.TransformComponentType)

	// Use the first active listener found
	for _, entityID := range listenerEntities {
		as.listenerEntity = entityID
		break
	}
}

// updateAudioSources updates all audio sources
func (as *AudioSystem) updateAudioSources(deltaTime float32) {
	as.activeAudioSources = as.activeAudioSources[:0]

	audioEntities := as.world.GetEntitiesWithComponents(components.AudioSourceComponentType, components.TransformComponentType)

	for _, entityID := range audioEntities {
		audioComp, _ := as.world.GetComponent(entityID, components.AudioSourceComponentType)
		transformComp, _ := as.world.GetComponent(entityID, components.TransformComponentType)

		if audioSource, ok := audioComp.(*components.AudioSourceComponent); ok {
			if transform, ok := transformComp.(*components.TransformComponent); ok {
				// Update fade effects
				audioSource.UpdateFade(deltaTime)

				// Update current time
				if audioSource.IsPlaying && !audioSource.IsPaused {
					audioSource.CurrentTime += deltaTime
				}

				// Check if sound should loop
				if audioSource.IsLooping && audioSource.IsPlaying && audioSource.AudioClipLength > 0 {
					if audioSource.CurrentTime >= audioSource.AudioClipLength {
						audioSource.CurrentTime = 0.0
						// Restart the sound
						rl.StopSound(audioSource.Sound)
						rl.PlaySound(audioSource.Sound)
					}
				}

				// Create active audio source entry
				activeSource := ActiveAudioSource{
					EntityID:     entityID,
					AudioSource:  audioSource,
					Transform:    transform,
					Priority:     audioSource.Priority,
					LastPosition: transform.Position,
				}

				// Calculate distance to listener
				if as.listenerEntity != 0 {
					if listenerTransformComp, exists := as.world.GetComponent(as.listenerEntity, components.TransformComponentType); exists {
						if listenerTransform, ok := listenerTransformComp.(*components.TransformComponent); ok {
							activeSource.Distance = core.Vector3Distance(transform.Position, listenerTransform.Position)
						}
					}
				}

				// Calculate effective volume
				activeSource.Volume = audioSource.GetEffectiveVolume(activeSource.Distance)
				activeSource.IsAudible = activeSource.Volume > 0.01 // Threshold for audibility

				as.activeAudioSources = append(as.activeAudioSources, activeSource)
			}
		}
	}

	// Sort by priority and distance for audio source management
	as.sortAudioSources()

	// Limit active audio sources
	as.limitAudioSources()
}

// sortAudioSources sorts audio sources by priority and distance
func (as *AudioSystem) sortAudioSources() {
	// Simple bubble sort by priority (higher first), then by distance (closer first)
	n := len(as.activeAudioSources)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			a := as.activeAudioSources[j]
			b := as.activeAudioSources[j+1]

			shouldSwap := false
			if a.Priority < b.Priority {
				shouldSwap = true
			} else if a.Priority == b.Priority && a.Distance > b.Distance {
				shouldSwap = true
			}

			if shouldSwap {
				as.activeAudioSources[j], as.activeAudioSources[j+1] = as.activeAudioSources[j+1], as.activeAudioSources[j]
			}
		}
	}
}

// limitAudioSources limits the number of simultaneously playing audio sources
func (as *AudioSystem) limitAudioSources() {
	if len(as.activeAudioSources) <= as.maxAudioSources {
		return
	}

	// Stop audio sources beyond the limit
	for i := as.maxAudioSources; i < len(as.activeAudioSources); i++ {
		audioSource := as.activeAudioSources[i].AudioSource
		if audioSource.IsPlaying {
			audioSource.Stop()
		}
	}

	// Keep only the highest priority sources
	as.activeAudioSources = as.activeAudioSources[:as.maxAudioSources]
}

// updateReverbZones updates reverb zone influences
func (as *AudioSystem) updateReverbZones() {
	as.reverbZones = as.reverbZones[:0]

	reverbEntities := as.world.GetEntitiesWithComponents(components.AudioReverbZoneComponentType, components.TransformComponentType)

	if as.listenerEntity == 0 {
		return
	}

	listenerTransformComp, exists := as.world.GetComponent(as.listenerEntity, components.TransformComponentType)
	if !exists {
		return
	}

	listenerTransform, ok := listenerTransformComp.(*components.TransformComponent)
	if !ok {
		return
	}

	for _, entityID := range reverbEntities {
		reverbComp, _ := as.world.GetComponent(entityID, components.AudioReverbZoneComponentType)
		transformComp, _ := as.world.GetComponent(entityID, components.TransformComponentType)

		if reverbZone, ok := reverbComp.(*components.AudioReverbZoneComponent); ok {
			if transform, ok := transformComp.(*components.TransformComponent); ok {
				if reverbZone.Enabled {
					distance := core.Vector3Distance(listenerTransform.Position, transform.Position)
					influence := as.calculateReverbInfluence(distance, reverbZone)

					if influence > 0.0 {
						as.reverbZones = append(as.reverbZones, ReverbZoneData{
							EntityID:   entityID,
							ReverbZone: reverbZone,
							Transform:  transform,
							Influence:  influence,
						})
					}
				}
			}
		}
	}
}

// calculateReverbInfluence calculates the influence of a reverb zone
func (as *AudioSystem) calculateReverbInfluence(distance float32, reverbZone *components.AudioReverbZoneComponent) float32 {
	if distance <= reverbZone.MinDistance {
		return 1.0
	} else if distance >= reverbZone.MaxDistance {
		return 0.0
	} else {
		// Linear falloff between min and max distance
		return 1.0 - (distance-reverbZone.MinDistance)/(reverbZone.MaxDistance-reverbZone.MinDistance)
	}
}

// process3DAudio processes 3D spatial audio effects
func (as *AudioSystem) process3DAudio(deltaTime float32) {
	if as.listenerEntity == 0 {
		return
	}

	listenerTransformComp, exists := as.world.GetComponent(as.listenerEntity, components.TransformComponentType)
	if !exists {
		return
	}

	listenerTransform, ok := listenerTransformComp.(*components.TransformComponent)
	if !ok {
		return
	}

	// Get listener component
	listenerComp, exists := as.world.GetComponent(as.listenerEntity, components.AudioListenerComponentType)
	var listener *components.AudioListenerComponent
	if exists {
		if l, ok := listenerComp.(*components.AudioListenerComponent); ok {
			listener = l
			listener.UpdateVelocity(listenerTransform.Position, deltaTime)
		}
	}

	// Process each active audio source
	for i := range as.activeAudioSources {
		source := &as.activeAudioSources[i]
		as.process3DAudioSource(source, listenerTransform, listener, deltaTime)
	}
}

// process3DAudioSource processes 3D audio for a single source
func (as *AudioSystem) process3DAudioSource(source *ActiveAudioSource, listenerTransform *components.TransformComponent, listener *components.AudioListenerComponent, deltaTime float32) {
	if !source.AudioSource.Is3D || source.AudioSource.SpatialBlend == 0.0 {
		// 2D audio - just apply volume
		source.AudioSource.SetVolume(source.AudioSource.Volume * as.masterVolume)
		return
	}

	// Calculate 3D audio parameters
	distance := source.Distance
	direction := core.Vector3Normalize(core.Vector3Subtract(source.Transform.Position, listenerTransform.Position))

	// Calculate volume based on distance
	volume := source.AudioSource.GetEffectiveVolume(distance) * as.masterVolume

	// Calculate Doppler effect if enabled
	if as.dopplerEnabled && listener != nil && source.AudioSource.DopplerFactor > 0.0 {
		pitch := as.calculateDopplerPitch(source, listenerTransform, listener, deltaTime)
		source.AudioSource.SetPitch(source.AudioSource.Pitch * pitch)
	}

	// Calculate stereo panning based on position
	if listener != nil {
		_ = as.calculateStereoPan(direction, listenerTransform)
		// Apply pan (raylib doesn't have direct pan control, so this would need custom implementation)
		// For now, we'll just adjust volume
	}

	// Apply final volume
	source.AudioSource.SetVolume(volume)

	// Update velocity for next frame (for Doppler)
	if deltaTime > 0 {
		displacement := core.Vector3Subtract(source.Transform.Position, source.LastPosition)
		source.Velocity = core.Vector3Scale(displacement, 1.0/deltaTime)
		source.LastPosition = source.Transform.Position
	}
}

// calculateDopplerPitch calculates the Doppler effect pitch multiplier
func (as *AudioSystem) calculateDopplerPitch(source *ActiveAudioSource, listenerTransform *components.TransformComponent, listener *components.AudioListenerComponent, deltaTime float32) float32 {
	if deltaTime == 0 || listener.SpeedOfSound == 0 {
		return 1.0
	}

	// Calculate relative velocity
	relativeVelocity := core.Vector3Subtract(source.Velocity, listener.Velocity)
	direction := core.Vector3Normalize(core.Vector3Subtract(source.Transform.Position, listenerTransform.Position))

	// Project relative velocity onto the line between source and listener
	velocityAlongLine := rl.Vector3DotProduct(relativeVelocity, direction)

	// Doppler formula: f' = f * (v + vr) / (v + vs)
	// where v = speed of sound, vr = listener velocity, vs = source velocity
	dopplerShift := (listener.SpeedOfSound - velocityAlongLine) / listener.SpeedOfSound

	// Apply Doppler level scaling
	dopplerShift = 1.0 + (dopplerShift - 1.0) * listener.DopplerLevel * source.AudioSource.DopplerFactor

	// Clamp to reasonable range
	if dopplerShift < 0.1 {
		dopplerShift = 0.1
	} else if dopplerShift > 3.0 {
		dopplerShift = 3.0
	}

	return dopplerShift
}

// calculateStereoPan calculates stereo panning based on audio source direction
func (as *AudioSystem) calculateStereoPan(direction rl.Vector3, listenerTransform *components.TransformComponent) float32 {
	// Get listener's right vector (simplified - assumes Y is up)
	listenerRight := core.Vector3Normalize(rl.Vector3{X: 1, Y: 0, Z: 0}) // Simplified

	// Calculate dot product to determine left/right position
	pan := rl.Vector3DotProduct(direction, listenerRight)

	// Clamp to [-1, 1] range
	if pan < -1.0 {
		pan = -1.0
	} else if pan > 1.0 {
		pan = 1.0
	}

	return pan
}

// updateSoundPlayback updates sound playback state
func (as *AudioSystem) updateSoundPlayback(deltaTime float32) {
	for _, source := range as.activeAudioSources {
		audioSource := source.AudioSource

		// Handle PlayOnAwake
		if audioSource.PlayOnAwake && !audioSource.IsPlaying && !audioSource.IsPaused {
			audioSource.Play()
			audioSource.PlayOnAwake = false // Only play once
		}

		// Check if sound has finished playing (for non-looping sounds)
		if audioSource.IsPlaying && !audioSource.IsLooping {
			// In a full implementation, you'd check if the sound has actually finished
			// raylib doesn't provide direct access to this, so we'd need to track it manually
			if audioSource.AudioClipLength > 0 && audioSource.CurrentTime >= audioSource.AudioClipLength {
				audioSource.Stop()
			}
		}
	}
}

// Configuration methods

// SetMasterVolume sets the master volume for all audio
func (as *AudioSystem) SetMasterVolume(volume float32) {
	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}
	as.masterVolume = volume
	rl.SetMasterVolume(volume)
}

// GetMasterVolume returns the current master volume
func (as *AudioSystem) GetMasterVolume() float32 {
	return as.masterVolume
}

// SetMaxAudioSources sets the maximum number of simultaneously playing audio sources
func (as *AudioSystem) SetMaxAudioSources(max int) {
	if max < 1 {
		max = 1
	} else if max > 64 {
		max = 64
	}
	as.maxAudioSources = max
}

// SetDistanceModel sets the distance model for 3D audio
func (as *AudioSystem) SetDistanceModel(model DistanceModel) {
	as.distanceModel = model
}

// SetDopplerEnabled enables or disables Doppler effect
func (as *AudioSystem) SetDopplerEnabled(enabled bool) {
	as.dopplerEnabled = enabled
}

// GetActiveAudioSourceCount returns the number of currently active audio sources
func (as *AudioSystem) GetActiveAudioSourceCount() int {
	return len(as.activeAudioSources)
}

// GetAudioDeviceInfo returns information about the audio device
func (as *AudioSystem) GetAudioDeviceInfo() (sampleRate, bufferSize, channels int) {
	return as.sampleRate, as.bufferSize, as.channels
}

// PlayOneShot plays a sound effect once at a specific position
func (as *AudioSystem) PlayOneShot(sound rl.Sound, position rl.Vector3, volume float32) {
	// Create a temporary entity for the one-shot sound
	entity := as.world.CreateEntity()

	// Add transform component
	transform := components.NewTransformComponentAt(position)
	entity.AddComponent(transform)

	// Add audio source component
	audioSource := components.NewAudioSourceComponent(sound)
	audioSource.Volume = volume
	audioSource.PlayOnAwake = true
	entity.AddComponent(audioSource)

	// The entity will be cleaned up when the sound finishes playing
	// In a full implementation, you'd want a cleanup system for temporary entities
}