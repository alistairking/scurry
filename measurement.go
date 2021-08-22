package scurry

import (
	"encoding/json"
	"fmt"
)

type MeasurementCommand interface {
	AsCommand() string
}

//go:generate enumer -type=MeasurementType -json -text -linecomment
type MeasurementType int

const (
	MEASUREMENT_UNKNOWN    MeasurementType = iota // unknown
	MEASUREMENT_PING                              // ping
	MEASUREMENT_TRACEROUTE                        // trace
)

type NoopOptions struct{}

func (n NoopOptions) AsCommand() string {
	return ""
}

type Ping struct {
	// TODO
}

func (p Ping) AsCommand() string {
	// TODO
	return ""
}

type Traceroute struct {
	// TODO
}

func (t Traceroute) AsCommand() string {
	// TODO
	return ""
}

type MeasurementOpts struct {
	// MeasurementType-specific config:
	Ping       Ping       `json:"ping"`
	Traceroute Traceroute `json:"traceroute"`
}

type Measurement struct {
	Type    MeasurementType `json:"type"`
	Target  string          `json:"target"`
	Options MeasurementOpts `json:"options"`

	Result ScResult `json:"result"`

	userId uint64 // used internally to match results with measurements
}

func (m Measurement) GetTypeOptions() MeasurementCommand {
	switch m.Type {
	case MEASUREMENT_PING:
		return m.Options.Ping
	case MEASUREMENT_TRACEROUTE:
		return m.Options.Traceroute
	}
	return NoopOptions{}
}

func (m Measurement) AsCommand() string {
	return fmt.Sprintf(
		"%s -U %d %s %s",
		m.Type.String(),
		m.userId,
		m.GetTypeOptions().AsCommand(),
		m.Target,
	)
}

func (m Measurement) AsJson() (string, error) {
	d, err := json.Marshal(m)
	return string(d), err
}

func (m Measurement) String() string {
	j, _ := m.AsJson()
	return j
}
