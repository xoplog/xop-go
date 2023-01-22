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
			request := logger.Request(ctx, event.Span.StartTime, event.Span.Bundle, event.Span.Name)
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
				// next line must be blank to end macro
				case xopbase.BoolDataType:
					line.Bool(k, v.(bool))
				// next line must be blank to end macro
				case xopbase.DurationDataType:
					line.Duration(k, v.(time.Duration))
				// next line must be blank to end macro
				case xopbase.TimeDataType:
					line.Time(k, v.(time.Time))
				// next line must be blank to end macro

				case xopbase.Float64DataType:
					line.Float64(k, v.(float64), dataType)
				// next line must be blank to end macro
				case xopbase.Int64DataType:
					line.Int64(k, v.(int64), dataType)
				// next line must be blank to end macro
				case xopbase.StringDataType:
					line.String(k, v.(string), dataType)
				// next line must be blank to end macro
				case xopbase.Uint64DataType:
					line.Uint64(k, v.(uint64), dataType)
				// next line must be blank to end macro

				case xopbase.EnumDataType:
					// XXX TODO
				default:
					return errors.Errorf("unexpected data type %s in line", dataType)
				}
			}
		case MetadataSet:
			span, ok := spans[event.Span.Bundle.Trace.GetSpanID()]
			if !ok {
				return errors.Errorf("missing span %s for metadataSet", event.Span.Bundle.Trace)
			}
			switch event.Attribute.SubType() {
			case xopat.AttributeTypeAny:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						span.MetadataAny(event.Attribute.(*xopat.AnyAttribute), v)
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					span.MetadataAny(event.Attribute.(*xopat.AnyAttribute), v)
				}
				// next line must be blank to end macro
			case xopat.AttributeTypeBool:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						span.MetadataBool(event.Attribute.(*xopat.BoolAttribute), v.(bool))
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					span.MetadataBool(event.Attribute.(*xopat.BoolAttribute), v.(bool))
				}
				// next line must be blank to end macro
			case xopat.AttributeTypeEnum:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						enum, ok := event.Attribute.GetEnum(v.(string))
						if !ok {
							return errors.Errorf("missing enum value for %s key %s", v.(string), event.Attribute.Key())
						}
						span.MetadataEnum(event.Attribute.(*xopat.EnumAttribute), enum)
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					enum, ok := event.Attribute.GetEnum(v.(string))
					if !ok {
						return errors.Errorf("missing enum value for %s key %s", v.(string), event.Attribute.Key())
					}
					span.MetadataEnum(event.Attribute.(*xopat.EnumAttribute), enum)
				}
				// next line must be blank to end macro
			case xopat.AttributeTypeFloat64:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						span.MetadataFloat64(event.Attribute.(*xopat.Float64Attribute), v.(float64))
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					span.MetadataFloat64(event.Attribute.(*xopat.Float64Attribute), v.(float64))
				}
				// next line must be blank to end macro
			case xopat.AttributeTypeInt64:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						span.MetadataInt64(event.Attribute.(*xopat.Int64Attribute), v.(int64))
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					span.MetadataInt64(event.Attribute.(*xopat.Int64Attribute), v.(int64))
				}
				// next line must be blank to end macro
			case xopat.AttributeTypeLink:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						span.MetadataLink(event.Attribute.(*xopat.LinkAttribute), v.(xoptrace.Trace))
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					span.MetadataLink(event.Attribute.(*xopat.LinkAttribute), v.(xoptrace.Trace))
				}
				// next line must be blank to end macro
			case xopat.AttributeTypeString:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						span.MetadataString(event.Attribute.(*xopat.StringAttribute), v.(string))
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					span.MetadataString(event.Attribute.(*xopat.StringAttribute), v.(string))
				}
				// next line must be blank to end macro
			case xopat.AttributeTypeTime:
				if event.Attribute.Multiple() {
					for _, v := range event.Span.Metadata[event.Attribute.Key()].([]interface{}) {
						span.MetadataTime(event.Attribute.(*xopat.TimeAttribute), v.(time.Time))
					}
				} else {
					v := event.Span.Metadata[event.Attribute.Key()]
					span.MetadataTime(event.Attribute.(*xopat.TimeAttribute), v.(time.Time))
				}
				// next line must be blank to end macro

			default:
				return errors.Errorf("unknown attribute type %s", event.Attribute.ProtoType())
			}
		default:

		}
	}
	return nil
}
