package xop

import (
	"time"

	"github.com/muir/xop/trace"
	"github.com/muir/xop/zap"
)

// BaseLogger is the bottom half of a logger -- the part that actually
// outputs data somewhere.  There can be many BaseLogger implementations.
type BaseLogger interface {
	Request() BaseRequest

	// ReferencesOkay should return true if references to objects are okay
	// as parameters to Any() and inside Things.  References are okay as long
	// as the objects are immediately encoded or transformed.
	ReferencesOkay() bool

	Close()
}

type BaseRequest interface {
	Flush()
	Span(span trace.Trace) BaseSpan
}

type BaseSpan interface {
	Line(Level, time.Time) BaseLine
	Data([]xopthing.Thing)
	LinePrefil([]xopthing.Thing) // adds
	ResetLinePrefill()
	Span(span trace.Trace) BaseSpan // inherits line prefill
}

type BaseLine interface {
	BaseObjectParts
	// TODO: ExternalReference(name string, itemId string, storageId string)
	Msg(string)
}

type SubObject interface {
	BaseObjectParts
	Complete()
}

type Encoder interface {
	MimeType() string
	ProducesText() bool
	Encode(elementName string, data interface{}) ([]byte, error)
}

type BaseObjectParts interface {
	Int(string, int64)
	Uint(string, uint64)
	Bool(string, bool)
	Str(string, string)
	Time(string, time.Time)
	Error(string, error)
	Any(string, interface{}) // generally serialized with JSON
	// TODO: SubObject(string) SubObject
	// TODO: Encoded(name string, elementName string, encoder Encoder, data interface{})
	// TODO: PreEncodedBytes(name string, elementName string, mimeType string, data []byte)
	// TODO: PreEncodedText(name string, elementName string, mimeType string, data string)
	// TODO: Guage()
	// TODO: Event()
}

type BaseBuffer interface {
	Context()
	Flush()
}
