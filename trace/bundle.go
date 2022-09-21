package trace

type Bundle struct {
	TraceParent Trace
	Trace       Trace
	State       State
	Baggage     Baggage
}

func NewBundle() Bundle {
	return Bundle{
		TraceParent: NewTrace(),
		Trace:       NewTrace(),
	}
}

func (b Bundle) Copy() Bundle {
	return Bundle{
		TraceParent: b.TraceParent.Copy(),
		Trace:       b.Trace.Copy(),
		State:       b.State.Copy(),
		Baggage:     b.Baggage.Copy(),
	}
}
