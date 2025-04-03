package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

// Store manages the storage of command presets
type Store struct {
	PresetList PresetList
	configFile string
}

// NewStore creates a new data store and loads existing presets
func NewStore() (*Store, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	configDir := filepath.Join(home, ".config", "command-presets")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "presets.json")
	store := &Store{
		PresetList: PresetList{Presets: []Preset{}},
		configFile: configFile,
	}

	if err := store.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load presets: %v", err)
	}

	return store, nil
}

// Load reads presets from the config file
func (s *Store) Load() error {
	data, err := ioutil.ReadFile(s.configFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.PresetList)
}

// Save writes presets to the config file
func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.PresetList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal presets: %v", err)
	}

	return ioutil.WriteFile(s.configFile, data, 0644)
}

// AddPreset adds a new preset to the store
func (s *Store) AddPreset(preset Preset) {
	s.PresetList.Presets = append(s.PresetList.Presets, preset)
}

// RemovePreset removes a preset by index
func (s *Store) RemovePreset(index int) error {
	if index < 0 || index >= len(s.PresetList.Presets) {
		return fmt.Errorf("invalid preset index")
	}

	s.PresetList.Presets = append(s.PresetList.Presets[:index], s.PresetList.Presets[index+1:]...)
	return nil
}

// ImportPresets imports presets from a JSON file
func (s *Store) ImportPresets(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var importedList PresetList
	if err := json.Unmarshal(data, &importedList); err != nil {
		return err
	}

	s.PresetList.Presets = append(s.PresetList.Presets, importedList.Presets...)
	return nil
}

// ExportPresets exports presets to a JSON file
func (s *Store) ExportPresets(filePath string) error {
	data, err := json.MarshalIndent(s.PresetList, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}