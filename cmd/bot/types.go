package main

type WatchMap map[int]int64
type StateMap map[int]string

type OffsetData struct {
	Offset int `json:"offset"`
}
