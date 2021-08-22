package scurry

import (
	"encoding/json"
)

// TODO: this is ping only, do traceroute too
type ScResult struct {
	Type      string `json:"type"`
	Version   string `json:"version"`
	Method    string `json:"method"`
	Src       string `json:"src"`
	Dst       string `json:"dst"`
	Start     ScTime `json:"start"`
	PingSent  int    `json:"ping_sent"`
	ProbeSize int    `json:"probe_size"`
	UserID    uint64 `json:"userid"`
	TTL       uint8  `json:"ttl"`
	Wait      int    `json:"wait"`
	Timeout   int    `json:"timeout"`
	// TODO responses
	// TODO statistics
}

func NewScResultFromJson(scJson string) (*ScResult, error) {
	var res ScResult
	if err := json.Unmarshal([]byte(scJson), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r ScResult) String() string {
	d, _ := json.Marshal(r)
	return string(d)
}

type ScTime struct {
	Sec  uint64 `json:"sec"`
	Usec uint64 `json:"usec"`
}
