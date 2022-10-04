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
