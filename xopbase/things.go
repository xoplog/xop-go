package xopbase

import (
	"fmt"
	"time"

	"github.com/muir/xoplog/xop"
)

func LineThings(line Line, things []xop.Thing) {
	for _, thing := range things {
		switch thing.Type {
		case xop.IntType:
			line.Int(thing.Key, thing.Int)
		case xop.UintType:
			line.Uint(thing.Key, thing.Any.(uint64))
		case xop.BoolType:
			line.Bool(thing.Key, thing.Any.(bool))
		case xop.StringType:
			line.Str(thing.Key, thing.String)
		case xop.TimeType:
			line.Time(thing.Key, thing.Any.(time.Time))
		case xop.AnyType:
			line.Any(thing.Key, thing.Any)
		case xop.ErrorType:
			line.Error(thing.Key, thing.Any.(error))
		case xop.UnsetType:
			fallthrough
		default:
			panic(fmt.Sprintf("malformed xop.Thing, type is %d", thing.Type))
		}
	}
}
