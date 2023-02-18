// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoptest

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptrace"
)

func (log *TestLogger) Replay(ctx context.Context, input any, logger xopbase.Logger) error {
	return log.LosslessReplay(ctx, input, logger)
}

func (_ *TestLogger) LosslessReplay(ctx context.Context, input any, logger xopbase.Logger) error {
	log, ok := input.(*TestLogger)
	if !ok {
		return errors.Errorf("xoptest Replay only supports *TestLogger")
	}
	requests := make(map[xoptrace.HexBytes8]xopbase.Request)
	spans := make(map[xoptrace.HexBytes8]xopbase.Span)
	for _, event := range log.Events {
		switch event.Type {
		case CustomEvent:
			// ignore
		case RequestStart:
			request := logger.Request(ctx, event.Span.StartTime, event.Span.Bundle, event.Span.Name, *event.Span.SourceInfo)
			id := event.Span.Bundle.Trace.GetSpanID()
			requests[id] = request
			spans[id] = request
		case RequestDone:
			if req, ok := requests[event.Span.Bundle.Trace.GetSpanID()]; ok {
				req.Done(time.Unix(0, event.Span.EndTime), event.Done)
			} else {
				return errors.Errorf("RequestDone event without corresponding RequestStart for %s", event.Span.Bundle.Trace)
			}
		case SpanDone:
			if span, ok := spans[event.Span.Bundle.Trace.GetSpanID()]; ok {
				span.Done(time.Unix(0, event.Span.EndTime), event.Done)
			} else {
				return errors.Errorf("SpanDone event without corresponding SpanStart for %s", event.Span.Bundle.Trace)
			}
		case FlushEvent:
			id := event.Span.Bundle.Trace.GetSpanID()
			if req, ok := requests[id]; ok {
				req.Flush()
			} else {
				return errors.Errorf("Flush for unknown req %s", event.Span.Bundle.Trace)
			}
		case SpanStart:
			if event.Span.Parent == nil {
				return errors.Errorf("Span w/o parent, %s", event.Span.Bundle.Trace)
			}
			if parent, ok := spans[event.Span.Parent.Bundle.Trace.GetSpanID()]; ok {
				span := parent.Span(ctx, event.Span.StartTime, event.Span.Bundle, event.Span.Name, event.Span.SequenceCode)
				spans[event.Span.Bundle.Trace.GetSpanID()] = span
			}
		case LineEvent:
			span, ok := spans[event.Line.Span.Bundle.Trace.GetSpanID()]
			if !ok {
				return errors.Errorf("missing span %s for line", event.Line.Span.Bundle.Trace)
			}
			line := span.NoPrefill().Line(event.Line.Level, event.Line.Timestamp, nil /* XXX TODO */)
			for k, v := range event.Line.Data {
				dataType := event.Line.DataType[k]
				switch dataType {
				case xopbase.AnyDataType:
					line.Any(k, v.(xopbase.ModelArg))
				// next line must be blank to end macro BaseDataWithoutType
				case xopbase.BoolDataType:
					line.Bool(k, v.(bool))
				// next line must be blank to end macro BaseDataWithoutType
				case xopbase.DurationDataType:
					line.Duration(k, v.(time.Duration))
				// next line must be blank to end macro BaseDataWithoutType
				case xopbase.TimeDataType:
					line.Time(k, v.(time.Time))
				// next line must be blank to end macro BaseDataWithoutType

				case xopbase.Float64DataType:
					line.Float64(k, v.(float64), dataType)
				// next line must be blank to end macro BaseDataWithType
				case xopbase.StringDataType:
					line.String(k, v.(string), dataType)
				// next line must be blank to end macro BaseDataWithType

				case xopbase.IntDataType:
					line.Int64(k, v.(int64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Int16DataType:
					line.Int64(k, v.(int64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Int32DataType:
					line.Int64(k, v.(int64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Int64DataType:
					line.Int64(k, v.(int64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Int8DataType:
					line.Int64(k, v.(int64), dataType)
				// next line must be blank to end macro Ints

				case xopbase.UintDataType:
					line.Uint64(k, v.(uint64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Uint16DataType:
					line.Uint64(k, v.(uint64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Uint32DataType:
					line.Uint64(k, v.(uint64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Uint64DataType:
					line.Uint64(k, v.(uint64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.Uint8DataType:
					line.Uint64(k, v.(uint64), dataType)
				// next line must be blank to end macro Ints
				case xopbase.UintptrDataType:
					line.Uint64(k, v.(uint64), dataType)
				// next line must be blank to end macro Ints

				case xopbase.ErrorDataType, xopbase.StringerDataType:
					line.String(k, v.(string), dataType)
				case xopbase.Float32DataType:
					line.Float64(k, v.(float64), dataType)
				case xopbase.EnumDataType:
					line.Enum(event.Line.Enums[k], v.(xopat.Enum))
				default:
					return errors.Errorf("unexpected data type %s in line", dataType)
				}
			}
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
				return errors.Errorf("missing span %s for metadataSet", event.Span.Bundle.Trace)
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
					return errors.Errorf("missing enum value for %T key %s", v, event.Attribute.Key())
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
				return errors.Errorf("unknown attribute type %s", event.Attribute.ProtoType())
			}
		default:

		}
	}
	return nil
}
