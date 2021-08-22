package measurement

import (
	"encoding/json"
	"fmt"
)

// Central measurement task object. Represents both a measurment
// request to be sent to scamper and the results once the measurment
// has been executed.
//
// At a minimum, the `Type`, and `Target` fields must be populated.
//
// Although the UserId field is publicly accessible, it will be
// overwritten by the Controller to allow scamper results to be
// matched against a Task.
type Task struct {
	Type    Type     `json:"type"`
	Target  string   `json:"target"`
	Options TaskOpts `json:"options"`

	Result *ScResult `json:"result"`

	UserId uint64 // used internally to match results with measurements
}

type ScCommand interface {
	AsCommand() string
}

type TaskOpts struct {
	// Type-specific config:
	Ping  Ping  `json:"ping"`
	Trace Trace `json:"trace"`
}

func (t Task) TypeOptions() ScCommand {
	switch t.Type {
	case TYPE_PING:
		return t.Options.Ping
	case TYPE_TRACE:
		return t.Options.Trace
	}
	return Noop{}
}

func (t Task) AsCommand() string {
	return fmt.Sprintf(
		"%s -U %d %s %s",
		t.Type.String(),
		t.UserId,
		t.TypeOptions().AsCommand(),
		t.Target,
	)
}

func (t Task) AsJson() (string, error) {
	d, err := json.Marshal(t)
	return string(d), err
}

func (t Task) String() string {
	j, _ := t.AsJson()
	return j
}
