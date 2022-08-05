// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoputil

import (
	"sync"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopconst"
)

func (a *AttributeBuilder) MetadataAny(k *xopconst.AnyAttribute, v interface{}) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if k.Multiple() {
		a.Anys[k.Key()] = append(a.Anys[k.Key()], v)
	} else {
		a.Any[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

func (a *AttributeBuilder) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.TimesSeen[k.Key()]; ok {
				if _, ok := seenMap[v.UnixNano()]; ok {
					return
				}
			} else {
				a.TimesSeen[k.Key()] = make(map[int64]struct{})
			}
			a.TimesSeen[k.Key()][v.UnixNano()] = struct{}{}
		}
		a.Times[k.Key()] = append(a.Times[k.Key()], v)
	} else {
		a.Time[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

func (a *AttributeBuilder) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	// TODO: when trace.Trace can be a map key, let this auto-generate
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.LinksSeen[k.Key()]; ok {
				if _, ok := seenMap[v.HeaderString()]; ok {
					return
				}
			} else {
				a.LinksSeen[k.Key()] = make(map[string]struct{})
			}
			a.LinksSeen[k.Key()][v.HeaderString()] = struct{}{}
		}
		a.Links[k.Key()] = append(a.Links[k.Key()], v)
	} else {
		a.Link[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

func (a AttributeBuilder) IsChanged() bool {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if len(a.Any) != 0 {
		for k := range a.Any {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Anys) != 0 {
		for k := range a.Anys {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Bool) != 0 {
		for k := range a.Bool {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Bools) != 0 {
		for k := range a.Bools {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Enum) != 0 {
		for k := range a.Enum {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Enums) != 0 {
		for k := range a.Enums {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Float64) != 0 {
		for k := range a.Float64 {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Float64s) != 0 {
		for k := range a.Float64s {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Int64) != 0 {
		for k := range a.Int64 {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Int64s) != 0 {
		for k := range a.Int64s {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Link) != 0 {
		for k := range a.Link {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Links) != 0 {
		for k := range a.Links {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.String) != 0 {
		for k := range a.String {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Strings) != 0 {
		for k := range a.Strings {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Time) != 0 {
		for k := range a.Time {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}
	if len(a.Times) != 0 {
		for k := range a.Times {
			if _, ok := a.Changed[k]; !ok {
				continue
			}
			return true
		}
	}

	return false
}

func (a AttributeBuilder) IsEmpty() bool {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if len(a.Any) != 0 {
		return false
	}
	if len(a.Anys) != 0 {
		return false
	}
	if len(a.Bool) != 0 {
		return false
	}
	if len(a.Bools) != 0 {
		return false
	}
	if len(a.Enum) != 0 {
		return false
	}
	if len(a.Enums) != 0 {
		return false
	}
	if len(a.Float64) != 0 {
		return false
	}
	if len(a.Float64s) != 0 {
		return false
	}
	if len(a.Int64) != 0 {
		return false
	}
	if len(a.Int64s) != 0 {
		return false
	}
	if len(a.Link) != 0 {
		return false
	}
	if len(a.Links) != 0 {
		return false
	}
	if len(a.String) != 0 {
		return false
	}
	if len(a.Strings) != 0 {
		return false
	}
	if len(a.Time) != 0 {
		return false
	}
	if len(a.Times) != 0 {
		return false
	}

	return true
}

func (a AttributeBuilder) Combine() map[string]interface{} {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	m := make(map[string]interface{})

	if len(a.Any) != 0 {
		for k, v := range a.Any {
			m[k] = v
		}
	}
	if len(a.Anys) != 0 {
		for k, v := range a.Anys {
			m[k] = v
		}
	}
	if len(a.Bool) != 0 {
		for k, v := range a.Bool {
			m[k] = v
		}
	}
	if len(a.Bools) != 0 {
		for k, v := range a.Bools {
			m[k] = v
		}
	}
	if len(a.Enum) != 0 {
		for k, v := range a.Enum {
			m[k] = v
		}
	}
	if len(a.Enums) != 0 {
		for k, v := range a.Enums {
			m[k] = v
		}
	}
	if len(a.Float64) != 0 {
		for k, v := range a.Float64 {
			m[k] = v
		}
	}
	if len(a.Float64s) != 0 {
		for k, v := range a.Float64s {
			m[k] = v
		}
	}
	if len(a.Int64) != 0 {
		for k, v := range a.Int64 {
			m[k] = v
		}
	}
	if len(a.Int64s) != 0 {
		for k, v := range a.Int64s {
			m[k] = v
		}
	}
	if len(a.Link) != 0 {
		for k, v := range a.Link {
			m[k] = v
		}
	}
	if len(a.Links) != 0 {
		for k, v := range a.Links {
			m[k] = v
		}
	}
	if len(a.String) != 0 {
		for k, v := range a.String {
			m[k] = v
		}
	}
	if len(a.Strings) != 0 {
		for k, v := range a.Strings {
			m[k] = v
		}
	}
	if len(a.Time) != 0 {
		for k, v := range a.Time {
			m[k] = v
		}
	}
	if len(a.Times) != 0 {
		for k, v := range a.Times {
			m[k] = v
		}
	}

	return m
}

// Reset is required before using zero-initialized AttributeBuilder
func (a *AttributeBuilder) Reset() {
	a.Lock.Lock()
	defer a.Lock.Unlock()

	a.Any = make(map[string]interface{})
	a.Anys = make(map[string][]interface{})
	a.Bool = make(map[string]bool)
	a.Bools = make(map[string][]bool)
	a.Enum = make(map[string]xopconst.Enum)
	a.Enums = make(map[string][]xopconst.Enum)
	a.Float64 = make(map[string]float64)
	a.Float64s = make(map[string][]float64)
	a.Int64 = make(map[string]int64)
	a.Int64s = make(map[string][]int64)
	a.Link = make(map[string]trace.Trace)
	a.Links = make(map[string][]trace.Trace)
	a.String = make(map[string]string)
	a.Strings = make(map[string][]string)
	a.Time = make(map[string]time.Time)
	a.Times = make(map[string][]time.Time)

	a.BoolsSeen = make(map[string]map[bool]struct{})
	a.EnumsSeen = make(map[string]map[xopconst.Enum]struct{})
	a.Float64sSeen = make(map[string]map[float64]struct{})
	a.Int64sSeen = make(map[string]map[int64]struct{})
	a.StringsSeen = make(map[string]map[string]struct{})

	a.LinksSeen = make(map[string]map[string]struct{})
	a.TimesSeen = make(map[string]map[int64]struct{})
	a.Changed = make(map[string]struct{})
}

func (a *AttributeBuilder) MetadataBool(k *xopconst.BoolAttribute, v bool) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.BoolsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				a.BoolsSeen[k.Key()] = make(map[bool]struct{})
			}
			a.BoolsSeen[k.Key()][v] = struct{}{}
		}
		a.Bools[k.Key()] = append(a.Bools[k.Key()], v)
	} else {
		a.Bool[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

func (a *AttributeBuilder) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.EnumsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				a.EnumsSeen[k.Key()] = make(map[xopconst.Enum]struct{})
			}
			a.EnumsSeen[k.Key()][v] = struct{}{}
		}
		a.Enums[k.Key()] = append(a.Enums[k.Key()], v)
	} else {
		a.Enum[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

func (a *AttributeBuilder) MetadataFloat64(k *xopconst.Float64Attribute, v float64) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.Float64sSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				a.Float64sSeen[k.Key()] = make(map[float64]struct{})
			}
			a.Float64sSeen[k.Key()][v] = struct{}{}
		}
		a.Float64s[k.Key()] = append(a.Float64s[k.Key()], v)
	} else {
		a.Float64[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

func (a *AttributeBuilder) MetadataInt64(k *xopconst.Int64Attribute, v int64) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.Int64sSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				a.Int64sSeen[k.Key()] = make(map[int64]struct{})
			}
			a.Int64sSeen[k.Key()][v] = struct{}{}
		}
		a.Int64s[k.Key()] = append(a.Int64s[k.Key()], v)
	} else {
		a.Int64[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

func (a *AttributeBuilder) MetadataString(k *xopconst.StringAttribute, v string) {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.StringsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				a.StringsSeen[k.Key()] = make(map[string]struct{})
			}
			a.StringsSeen[k.Key()][v] = struct{}{}
		}
		a.Strings[k.Key()] = append(a.Strings[k.Key()], v)
	} else {
		a.String[k.Key()] = v
	}
	a.Changed[k.Key()] = struct{}{}
}

type AttributeBuilder struct {
	Any      map[string]interface{}
	Anys     map[string][]interface{}
	Bool     map[string]bool
	Bools    map[string][]bool
	Enum     map[string]xopconst.Enum
	Enums    map[string][]xopconst.Enum
	Float64  map[string]float64
	Float64s map[string][]float64
	Int64    map[string]int64
	Int64s   map[string][]int64
	Link     map[string]trace.Trace
	Links    map[string][]trace.Trace
	String   map[string]string
	Strings  map[string][]string
	Time     map[string]time.Time
	Times    map[string][]time.Time

	BoolsSeen    map[string]map[bool]struct{}
	EnumsSeen    map[string]map[xopconst.Enum]struct{}
	Float64sSeen map[string]map[float64]struct{}
	Int64sSeen   map[string]map[int64]struct{}
	StringsSeen  map[string]map[string]struct{}

	LinksSeen map[string]map[string]struct{}
	TimesSeen map[string]map[int64]struct{}
	Changed   map[string]struct{}
	Lock      sync.Mutex
}
