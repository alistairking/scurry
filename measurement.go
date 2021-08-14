package scurry

//go:generate enumer -type=MeasurementType -json -text -linecomment
type MeasurementType int

const (
	MEASUREMENT_UNKNOWN    MeasurementType = iota // unknown
	MEASUREMENT_PING                              // ping
	MEASUREMENT_TRACEROUTE                        // traceroute
)

type Ping struct {
	// TODO
}

type Measurement struct {
	Type   MeasurementType
	Target string

	// MeasurementType-specific config:
	Ping Ping
	// TODO: Traceroute
}
