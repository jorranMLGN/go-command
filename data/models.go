package data

// Preset represents a command preset with name, command, and working directory
type Preset struct {
	Name       string `json:"name"`
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
	Enabled    bool   `json:"enabled"`
}

// PresetList is a collection of command presets
type PresetList struct {
	Presets []Preset `json:"presets"`
}
