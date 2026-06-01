package main

import (
	"encoding/json"
	"os"
)

func LoadWatch() WatchMap {

	var watch WatchMap

	b, err := os.ReadFile(
		"watch.json",
	)

	if err != nil {
		return WatchMap{}
	}

	json.Unmarshal(
		b,
		&watch,
	)

	return watch
}

func SaveWatch(
	watch WatchMap,
) {

	b, _ := json.MarshalIndent(
		watch,
		"",
		"  ",
	)

	os.WriteFile(
		"watch.json",
		b,
		0644,
	)
}
