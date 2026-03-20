// Package profile handles player profiles with local storage.
package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Stats holds player statistics.
type Stats struct {
	GamesPlayed int     `json:"games_played"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	Draws       int     `json:"draws"`
	Score       float64 `json:"score"` // Total score (win=1, draw=0.5, loss=0)
}

// Profile represents a player profile.
type Profile struct {
	Name  string `json:"name"`
	Stats Stats  `json:"stats"`
}

// ProfileStore manages player profiles with local file storage.
type ProfileStore struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
	filePath string
}

// TimeControl represents a time control option.
type TimeControl struct {
	Name     string
	Duration int // Duration in minutes, 0 means unlimited
}

// Available time controls.
var TimeControls = []TimeControl{
	{Name: "Unlimited", Duration: 0},
	{Name: "1 min (Bullet)", Duration: 1},
	{Name: "3 min (Blitz)", Duration: 3},
	{Name: "5 min (Blitz)", Duration: 5},
	{Name: "10 min (Rapid)", Duration: 10},
	{Name: "20 min (Rapid)", Duration: 20},
	{Name: "60 min (Classical)", Duration: 60},
}

// NewProfileStore creates a new profile store with local file storage.
func NewProfileStore() (*ProfileStore, error) {
	// Get user's home directory for local storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create .chess directory if it doesn't exist
	chessDir := filepath.Join(homeDir, ".chess")
	if err := os.MkdirAll(chessDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(chessDir, "profiles.json")

	store := &ProfileStore{
		profiles: make(map[string]*Profile),
		filePath: filePath,
	}

	// Load existing profiles
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// load reads profiles from the local file.
func (s *ProfileStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.profiles)
}

// save writes profiles to the local file.
func (s *ProfileStore) save() error {
	data, err := json.MarshalIndent(s.profiles, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// GetProfile returns a profile by name, creating it if it doesn't exist.
func (s *ProfileStore) GetProfile(name string) *Profile {
	s.mu.RLock()
	profile, exists := s.profiles[name]
	s.mu.RUnlock()

	if exists {
		return profile
	}

	// Create new profile
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check after acquiring write lock
	if profile, exists = s.profiles[name]; exists {
		return profile
	}

	profile = &Profile{
		Name: name,
		Stats: Stats{
			GamesPlayed: 0,
			Wins:        0,
			Losses:      0,
			Draws:       0,
			Score:       0,
		},
	}
	s.profiles[name] = profile
	s.save()

	return profile
}

// GetAllProfiles returns all stored profiles.
func (s *ProfileStore) GetAllProfiles() []*Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profiles := make([]*Profile, 0, len(s.profiles))
	for _, p := range s.profiles {
		profiles = append(profiles, p)
	}
	return profiles
}

// GetProfileNames returns all profile names.
func (s *ProfileStore) GetProfileNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.profiles))
	for name := range s.profiles {
		names = append(names, name)
	}
	return names
}

// RecordGameResult updates statistics for both players based on game result.
// winner: "white", "black", or "draw"
func (s *ProfileStore) RecordGameResult(whitePlayer, blackPlayer, winner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	white := s.getOrCreateProfileLocked(whitePlayer)
	black := s.getOrCreateProfileLocked(blackPlayer)

	white.Stats.GamesPlayed++
	black.Stats.GamesPlayed++

	switch winner {
	case "white":
		white.Stats.Wins++
		white.Stats.Score += 1
		black.Stats.Losses++
	case "black":
		black.Stats.Wins++
		black.Stats.Score += 1
		white.Stats.Losses++
	case "draw":
		white.Stats.Draws++
		black.Stats.Draws++
		white.Stats.Score += 0.5
		black.Stats.Score += 0.5
	}

	return s.save()
}

// getOrCreateProfileLocked gets or creates a profile (must be called with lock held).
func (s *ProfileStore) getOrCreateProfileLocked(name string) *Profile {
	if profile, exists := s.profiles[name]; exists {
		return profile
	}

	profile := &Profile{
		Name: name,
		Stats: Stats{
			GamesPlayed: 0,
			Wins:        0,
			Losses:      0,
			Draws:       0,
			Score:       0,
		},
	}
	s.profiles[name] = profile
	return profile
}

// DeleteProfile removes a profile from the store.
func (s *ProfileStore) DeleteProfile(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.profiles, name)
	return s.save()
}
