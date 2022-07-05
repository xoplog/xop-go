package trace

type Bundle struct {
	ParentTrace Trace
	Trace       Trace
	State       State
	Baggage     Baggage
}

func (b Bundle) Copy() Bundle {
	return Bundle{
		ParentTrace: b.ParentTrace.Copy(),
		Trace:       b.Trace.Copy(),
		State:       b.State,
		Baggage:     b.Baggage,
	}
}
