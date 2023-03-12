package xoputil

// DecodeEnum is used for decoding metadata enums
type DecodeEnum struct {
	I int64  `json:"i"`
	S string `json:"s"`
}

func (e DecodeEnum) String() string { return e.S }
func (e DecodeEnum) Int64() int64   { return e.I }
