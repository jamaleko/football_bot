package main

import (
	"encoding/json"
	"os"
)

func LoadOffset() int {

	var data OffsetData

	b, err := os.ReadFile(
		"offset.json",
	)

	if err != nil {
		return 0
	}

	json.Unmarshal(
		b,
		&data,
	)

	return data.Offset
}

func SaveOffset(
	offset int,
) {

	data := OffsetData{
		Offset: offset,
	}

	b, _ := json.MarshalIndent(
		data,
		"",
		"  ",
	)

	os.WriteFile(
		"offset.json",
		b,
		0644,
	)
}
