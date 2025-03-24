package types

import "fmt"

// Mode 播放模式
type Mode uint8

const (
	PmUnknown Mode = iota
	PmListLoop
	PmOrdered
	PmSingleLoop
	PmListRandom
	PmInfRandom
	NormalPmLength
	PmIntelligent
)

type State uint8

const (
	Unknown State = iota
	Playing
	Paused
	Stopped
	Interrupted
)

func (s State) String() string {
	switch s {
	case Unknown:
		return "Unknown"
	case Playing:
		return "Playing"
	case Paused:
		return "Paused"
	case Stopped:
		return "Stopped"
	case Interrupted:
		return "Interrupted"
	default:
		return fmt.Sprintf("State(%d)", s) // Handle unknown states gracefully
	}
}
