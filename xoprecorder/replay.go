// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoprecorder

import (
	"context"
	"fmt"
	"time"

	"github.com/muir/gwrap"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptrace"
)

// Replay dumps the recorded logs to another base logger
func (log *Logger) Replay(ctx context.Context, dest xopbase.Logger) error {
	requests := make(map[xoptrace.HexBytes8]xopbase.Request)
	spans := make(map[xoptrace.HexBytes8]xopbase.Span)
	var downStreamError gwrap.AtomicValue[error]
	errorReporter := func(err error) {
		if err != nil {
			downStreamError.Store(err)
		}
	}
	for _, event := range log.Events {
		switch event.Type {
		case CustomEvent:
			// ignore
		case RequestStart:
			request := dest.Request(ctx, event.Span.StartTime, event.Span.Bundle, event.Span.Name, *event.Span.SourceInfo)
			request.SetErrorReporter(errorReporter)
			id := event.Span.Bundle.Trace.GetSpanID()
			requests[id] = request
			spans[id] = request
		case RequestDone:
			if req, ok := requests[event.Span.Bundle.Trace.GetSpanID()]; ok {
				req.Done(time.Unix(0, event.Span.EndTime), event.Done)
			} else {
				return fmt.Errorf("RequestDone event without corresponding RequestStart for %s", event.Span.Bundle.Trace)
			}
		case SpanDone:
			if span, ok := spans[event.Span.Bundle.Trace.GetSpanID()]; ok {
				span.Done(time.Unix(0, event.Span.EndTime), event.Done)
			} else {
				return fmt.Errorf("SpanDone event without corresponding SpanStart for %s", event.Span.Bundle.Trace)
			}
		case FlushEvent:
			id := event.Span.Bundle.Trace.GetSpanID()
			if req, ok := requests[id]; ok {
				req.Flush()
			} else {
				return fmt.Errorf("Flush for unknown req %s", event.Span.Bundle.Trace)
			}
		case SpanStart:
			if event.Span.Parent == nil {
				return fmt.Errorf("Span w/o parent, %s", event.Span.Bundle.Trace)
			}
			if parent, ok := spans[event.Span.Parent.Bundle.Trace.GetSpanID()]; ok {
				span := parent.Span(ctx, event.Span.StartTime, event.Span.Bundle, event.Span.Name, event.Span.SpanSequenceCode)
				spans[event.Span.Bundle.Trace.GetSpanID()] = span
			}
		case LineEvent:
			span, ok := spans[event.Line.Span.Bundle.Trace.GetSpanID()]
			if !ok {
				return fmt.Errorf("missing span %s for line", event.Line.Span.Bundle.Trace)
			}
			line := span.NoPrefill().Line(event.Line.Level, event.Line.Timestamp, event.Line.Stack)
			ReplayLineData(event.Line, line)
			switch {
			case event.Line.Tmpl != "":
				line.Template(event.Line.Tmpl)
			case event.Line.AsLink != nil:
				line.Link(event.Line.Message, *event.Line.AsLink)
			case event.Line.AsModel != nil:
				line.Model(event.Line.Message, *event.Line.AsModel)
			default:
				line.Msg(event.Line.Message)
			}
		case MetadataSet:
			span, ok := spans[event.Span.Bundle.Trace.GetSpanID()]
			if !ok {
				return fmt.Errorf("missing span %s for metadataSet", event.Span.Bundle.Trace)
			}
			switch event.Attribute.SubType().SpanAttributeType() {
			case xopat.AttributeTypeAny:
				v := event.Value
				span.MetadataAny(event.Attribute.(*xopat.AnyAttribute), v.(xopbase.ModelArg))
				// next line must be blank to end macro
			case xopat.AttributeTypeBool:
				v := event.Value
				span.MetadataBool(event.Attribute.(*xopat.BoolAttribute), v.(bool))
				// next line must be blank to end macro
			case xopat.AttributeTypeEnum:
				v := event.Value
				enum, ok := v.(xopat.Enum)
				if !ok {
					return fmt.Errorf("missing enum value for %T key %s", v, event.Attribute.Key())
				}
				span.MetadataEnum(event.Attribute.(*xopat.EnumAttribute), enum)
				// next line must be blank to end macro
			case xopat.AttributeTypeFloat64:
				v := event.Value
				span.MetadataFloat64(event.Attribute.(*xopat.Float64Attribute), v.(float64))
				// next line must be blank to end macro
			case xopat.AttributeTypeInt64:
				v := event.Value
				span.MetadataInt64(event.Attribute.(*xopat.Int64Attribute), v.(int64))
				// next line must be blank to end macro
			case xopat.AttributeTypeLink:
				v := event.Value
				span.MetadataLink(event.Attribute.(*xopat.LinkAttribute), v.(xoptrace.Trace))
				// next line must be blank to end macro
			case xopat.AttributeTypeString:
				v := event.Value
				span.MetadataString(event.Attribute.(*xopat.StringAttribute), v.(string))
				// next line must be blank to end macro
			case xopat.AttributeTypeTime:
				v := event.Value
				span.MetadataTime(event.Attribute.(*xopat.TimeAttribute), v.(time.Time))
				// next line must be blank to end macro

			default:
				return fmt.Errorf("unknown attribute type %s", event.Attribute.ProtoType())
			}
		default:
			return fmt.Errorf("unexpected event type %v", event.Type)
		}
	}
	return downStreamError.Load()
}

func ReplayLineData(source *Line, dest xopbase.Builder) {
	for key, v := range source.Data {
		dataType := source.DataType[key]
		switch dataType {
		case xopbase.AnyDataType:
			dest.Any(key, v.(xopbase.ModelArg))
		// next line must be blank to end macro BaseDataWithoutType
		case xopbase.BoolDataType:
			dest.Bool(key, v.(bool))
		// next line must be blank to end macro BaseDataWithoutType
		case xopbase.DurationDataType:
			dest.Duration(key, v.(time.Duration))
		// next line must be blank to end macro BaseDataWithoutType
		case xopbase.TimeDataType:
			dest.Time(key, v.(time.Time))
		// next line must be blank to end macro BaseDataWithoutType

		case xopbase.Float64DataType:
			dest.Float64(key, v.(float64), dataType)
		// next line must be blank to end macro BaseDataWithType
		case xopbase.StringDataType:
			dest.String(key, v.(string), dataType)
		// next line must be blank to end macro BaseDataWithType

		case xopbase.IntDataType:
			dest.Int64(key, v.(int64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Int16DataType:
			dest.Int64(key, v.(int64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Int32DataType:
			dest.Int64(key, v.(int64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Int64DataType:
			dest.Int64(key, v.(int64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Int8DataType:
			dest.Int64(key, v.(int64), dataType)
		// next line must be blank to end macro Ints

		case xopbase.UintDataType:
			dest.Uint64(key, v.(uint64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Uint16DataType:
			dest.Uint64(key, v.(uint64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Uint32DataType:
			dest.Uint64(key, v.(uint64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Uint64DataType:
			dest.Uint64(key, v.(uint64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.Uint8DataType:
			dest.Uint64(key, v.(uint64), dataType)
		// next line must be blank to end macro Ints
		case xopbase.UintptrDataType:
			dest.Uint64(key, v.(uint64), dataType)
		// next line must be blank to end macro Ints

		case xopbase.ErrorDataType, xopbase.StringerDataType:
			dest.String(key, v.(string), dataType)
		case xopbase.Float32DataType:
			dest.Float64(key, v.(float64), dataType)
		case xopbase.EnumDataType:
			dest.Enum(source.Enums[key], v.(xopat.Enum))
		default:
			dest.String(key, fmt.Sprintf("unexpected data type %s in line, with value of type %T: %+v", dataType, v, v), xopbase.ErrorDataType)
		}
	}
}
