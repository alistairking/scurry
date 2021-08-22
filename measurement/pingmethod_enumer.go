// Code generated by "enumer -type=PingMethod -json -text -linecomment"; DO NOT EDIT.

//
package measurement

import (
	"encoding/json"
	"fmt"
)

const _PingMethodName = "icmp-echoicmp-timetcp-syntcp-acktcp-ack-sporttcp-synacktcp-rstudpudp-dport"

var _PingMethodIndex = [...]uint8{0, 9, 18, 25, 32, 45, 55, 62, 65, 74}

func (i PingMethod) String() string {
	if i >= PingMethod(len(_PingMethodIndex)-1) {
		return fmt.Sprintf("PingMethod(%d)", i)
	}
	return _PingMethodName[_PingMethodIndex[i]:_PingMethodIndex[i+1]]
}

var _PingMethodValues = []PingMethod{0, 1, 2, 3, 4, 5, 6, 7, 8}

var _PingMethodNameToValueMap = map[string]PingMethod{
	_PingMethodName[0:9]:   0,
	_PingMethodName[9:18]:  1,
	_PingMethodName[18:25]: 2,
	_PingMethodName[25:32]: 3,
	_PingMethodName[32:45]: 4,
	_PingMethodName[45:55]: 5,
	_PingMethodName[55:62]: 6,
	_PingMethodName[62:65]: 7,
	_PingMethodName[65:74]: 8,
}

// PingMethodString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func PingMethodString(s string) (PingMethod, error) {
	if val, ok := _PingMethodNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to PingMethod values", s)
}

// PingMethodValues returns all values of the enum
func PingMethodValues() []PingMethod {
	return _PingMethodValues
}

// IsAPingMethod returns "true" if the value is listed in the enum definition. "false" otherwise
func (i PingMethod) IsAPingMethod() bool {
	for _, v := range _PingMethodValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for PingMethod
func (i PingMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for PingMethod
func (i *PingMethod) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("PingMethod should be a string, got %s", data)
	}

	var err error
	*i, err = PingMethodString(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface for PingMethod
func (i PingMethod) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for PingMethod
func (i *PingMethod) UnmarshalText(text []byte) error {
	var err error
	*i, err = PingMethodString(string(text))
	return err
}