// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopbaseutil

import (
	"sync"
	"time"

	"github.com/muir/gwrap"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptrace"
)

// SpanMetadata
type SpanMetadata struct {
	Map gwrap.SyncMap[string, *MetadataTracker]
}

type MetadataTracker struct {
	Mu        sync.Mutex
	Value     any
	Type      xopbase.DataType
	Attribute xopat.AttributeInterface
	Seen      any
}

func (s SpanMetadata) GetValue(k string) any {
	mt, loaded := s.Map.Load(k)
	if !loaded {
		return nil
	}
	return mt.Value
}

func (s SpanMetadata) Get(k string) *MetadataTracker {
	mt, loaded := s.Map.Load(k)
	if !loaded {
		return nil
	}
	return mt
}

// MetadataAny is a required method for xopbase.Span
func (s *SpanMetadata) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
	v.Encode()
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			var key string
			v.Encode()
			key = string(v.Encoded)
			if !loaded {
				seen := make(map[string]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[string]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.AnyArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.AnyDataType
		}
		tracker.Value = v
	}
}

// MetadataBool is a required method for xopbase.Span
func (s *SpanMetadata) MetadataBool(k *xopat.BoolAttribute, v bool) {
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			key := value
			if !loaded {
				seen := make(map[bool]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[bool]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.BoolArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.BoolDataType
		}
		tracker.Value = v
	}
}

// MetadataEnum is a required method for xopbase.Span
func (s *SpanMetadata) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			key := v.String()
			if !loaded {
				seen := make(map[string]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[string]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.EnumArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.EnumDataType
		}
		tracker.Value = v
	}
}

// MetadataFloat64 is a required method for xopbase.Span
func (s *SpanMetadata) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			key := value
			if !loaded {
				seen := make(map[float64]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[float64]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.Float64ArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.Float64DataType
		}
		tracker.Value = v
	}
}

// MetadataInt64 is a required method for xopbase.Span
func (s *SpanMetadata) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			key := value
			if !loaded {
				seen := make(map[int64]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[int64]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.Int64ArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.Int64DataType
		}
		tracker.Value = v
	}
}

// MetadataLink is a required method for xopbase.Span
func (s *SpanMetadata) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			key := value
			if !loaded {
				seen := make(map[xoptrace.Trace]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[xoptrace.Trace]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.LinkArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.LinkDataType
		}
		tracker.Value = v
	}
}

// MetadataString is a required method for xopbase.Span
func (s *SpanMetadata) MetadataString(k *xopat.StringAttribute, v string) {
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			key := value
			if !loaded {
				seen := make(map[string]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[string]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.StringArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.StringDataType
		}
		tracker.Value = v
	}
}

// MetadataTime is a required method for xopbase.Span
func (s *SpanMetadata) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	tracker, loaded := s.Map.Load(k.Key().String())
	if loaded {
		tracker.Mu.Lock()
	} else {
		n := &MetadataTracker{}
		n.Mu.Lock()
		tracker, loaded = s.Map.LoadOrStore(k.Key().String(), n)
		if loaded {
			tracker.Mu.Lock()
		} else {
			tracker.Attribute = k
		}
	}
	defer tracker.Mu.Unlock()
	if tracker.Attribute.Multiple() {
		value := v
		if tracker.Attribute.Distinct() {
			key := value
			if !loaded {
				seen := make(map[time.Time]struct{})
				tracker.Seen = seen
				seen[key] = struct{}{}
			} else {
				seen := tracker.Seen.(map[time.Time]struct{})
				if _, ok := seen[key]; ok {
					return
				}
				seen[key] = struct{}{}
			}
		}
		if loaded {
			tracker.Value = append(tracker.Value.([]any), value)
		} else {
			tracker.Value = []any{value}
			tracker.Type = xopbase.TimeArrayDataType
		}
	} else {
		if loaded {
			if tracker.Attribute.Locked() {
				return
			}
		} else {
			tracker.Type = xopbase.TimeDataType
		}
		tracker.Value = v
	}
}
