package trace

type Bundle struct {
	Parent  Trace
	Trace   Trace
	State   State
	Baggage Baggage
}

func NewBundle() Bundle {
	return Bundle{
		Parent: NewTrace(),
		Trace:  NewTrace(),
	}
}

func (b Bundle) Copy() Bundle {
	return Bundle{
		Parent:  b.Parent.Copy(),
		Trace:   b.Trace.Copy(),
		State:   b.State.Copy(),
		Baggage: b.Baggage.Copy(),
	}
}

// ParentTraceIsDifferent returns true if the Parent.TraceID is set
// and it's different from Trace.TraceID
func (b Bundle) ParentTraceIsDifferent() bool {
	return !b.Parent.traceID.IsZero() &&
		b.Parent.traceID.b != b.Trace.traceID.b
}
