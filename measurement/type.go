package measurement

//go:generate enumer -type=Type -json -text -linecomment
type Type int

const (
	TYPE_UNKNOWN Type = iota // unknown
	TYPE_PING                // ping
	TYPE_TRACE               // trace
)
