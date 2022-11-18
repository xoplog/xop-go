package xoptrace

// Baggage tracks the contents of key/values that are passed
// through a trace in the "baggage" header.
// See https://w3c.github.io/baggage/
// Note that baggage values may contain PII and should not be logged
// where PII isn't allowed.
//
// TODO: allow baggage to be used and modified
type Baggage struct {
	// TODO: allow bagggage to be manipulated
	// baggage []struct {
	// 	key    string   // printable except ()<>@,;:\"/[]?={}
	// 	values []string // no whitespace, double-quote, comman, semicolon or backslash
	// }
	asString string
}

func (b *Baggage) SetString(h string) { b.asString = h }
func (b Baggage) IsZero() bool        { return b.asString == "" }
func (b Baggage) String() string      { return b.asString }
func (b Baggage) Copy() Baggage       { return b }
func (b Baggage) Bytes() []byte       { return []byte(b.asString) } // TODO: improve performance
