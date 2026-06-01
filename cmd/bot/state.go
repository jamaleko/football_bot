package main

import (
	"encoding/json"
	"os"
)

func LoadState() StateMap {

	var state StateMap

	b, err := os.ReadFile(
		"state.json",
	)

	if err != nil {
		return StateMap{}
	}

	json.Unmarshal(
		b,
		&state,
	)

	return state
}

func SaveState(
	state StateMap,
) {

	b, _ := json.MarshalIndent(
		state,
		"",
		"  ",
	)

	os.WriteFile(
		"state.json",
		b,
		0644,
	)
}
