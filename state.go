package xm

// State tracks the contents of key/values that are passed
// through a trace in the "tracestate" header.
//
// TODO: allow state to be used and modified
type State struct {
	// TODO: allow state to be manipulated
	// 	state       []struct {
	// 		vendor string // must match `\a(?:[a-z][-a-z0-9_*/]{0,255}|[a-z0-9][-a-z0-9_*/]{0,240}@[a-z]{1,14})\z`
	// 		key    string // upto 256 characters, all non-whitespace printable ascii except "," and "="
	// 	}
	asString string
}

func (s Seed) State() *State        { return &s.traceState.state }
func (s *State) SetString(h string) { s.asString = h }
func (s State) IsZero() bool        { return s.asString == "" }
func (s State) String() string      { return s.asString }
func (s State) Copy() State         { return s }
