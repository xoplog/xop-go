package trace

type Bundle struct {
	ParentTrace Trace
	Trace       Trace
	State       State
	Baggage     Baggage
}

func NewBundle() Bundle {
	return Bundle{
		ParentTrace: NewTrace(),
		Trace:       NewTrace(),
	}
}

func (b Bundle) Copy() Bundle {
	return Bundle{
		ParentTrace: b.ParentTrace.Copy(),
		Trace:       b.Trace.Copy(),
		State:       b.State.Copy(),
		Baggage:     b.Baggage.Copy(),
	}
}
