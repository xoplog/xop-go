// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoputil

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopconst"
)

func (a *AttributeBuilder) MetadataAny(k *xopconst.AnyAttribute, v interface{}) {
	if k.Multiple() {
		a.Anys[k.Key()] = append(a.Anys[k.Key()], v)
	} else {
		a.Any[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
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
}

func (a *AttributeBuilder) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
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
}

func (a AttributeBuilder) Combine() map[string]interface{} {
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
	if len(a.Number) != 0 {
		for k, v := range a.Number {
			m[k] = v
		}
	}
	if len(a.Numbers) != 0 {
		for k, v := range a.Numbers {
			m[k] = v
		}
	}
	if len(a.Str) != 0 {
		for k, v := range a.Str {
			m[k] = v
		}
	}
	if len(a.Strs) != 0 {
		for k, v := range a.Strs {
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
	a.Any = make(map[string]interface{})
	a.Anys = make(map[string][]interface{})
	a.Bool = make(map[string]bool)
	a.Bools = make(map[string][]bool)
	a.Enum = make(map[string]xopconst.Enum)
	a.Enums = make(map[string][]xopconst.Enum)
	a.Int64 = make(map[string]int64)
	a.Int64s = make(map[string][]int64)
	a.Link = make(map[string]trace.Trace)
	a.Links = make(map[string][]trace.Trace)
	a.Number = make(map[string]float64)
	a.Numbers = make(map[string][]float64)
	a.Str = make(map[string]string)
	a.Strs = make(map[string][]string)
	a.Time = make(map[string]time.Time)
	a.Times = make(map[string][]time.Time)

	a.BoolsSeen = make(map[string]map[bool]struct{})
	a.EnumsSeen = make(map[string]map[xopconst.Enum]struct{})
	a.Int64sSeen = make(map[string]map[int64]struct{})
	a.NumbersSeen = make(map[string]map[float64]struct{})
	a.StrsSeen = make(map[string]map[string]struct{})

	a.LinksSeen = make(map[string]map[string]struct{})
	a.TimesSeen = make(map[string]map[int64]struct{})
}

func (a *AttributeBuilder) MetadataBool(k *xopconst.BoolAttribute, v bool) {
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
}

func (a *AttributeBuilder) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
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
}

func (a *AttributeBuilder) MetadataInt64(k *xopconst.Int64Attribute, v int64) {
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
}

func (a *AttributeBuilder) MetadataNumber(k *xopconst.NumberAttribute, v float64) {
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.NumbersSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				a.NumbersSeen[k.Key()] = make(map[float64]struct{})
			}
			a.NumbersSeen[k.Key()][v] = struct{}{}
		}
		a.Numbers[k.Key()] = append(a.Numbers[k.Key()], v)
	} else {
		a.Number[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataStr(k *xopconst.StrAttribute, v string) {
	if k.Multiple() {
		if k.Distinct() {
			if seenMap, ok := a.StrsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				a.StrsSeen[k.Key()] = make(map[string]struct{})
			}
			a.StrsSeen[k.Key()][v] = struct{}{}
		}
		a.Strs[k.Key()] = append(a.Strs[k.Key()], v)
	} else {
		a.Str[k.Key()] = v
	}
}

type AttributeBuilder struct {
	Any     map[string]interface{}
	Anys    map[string][]interface{}
	Bool    map[string]bool
	Bools   map[string][]bool
	Enum    map[string]xopconst.Enum
	Enums   map[string][]xopconst.Enum
	Int64   map[string]int64
	Int64s  map[string][]int64
	Link    map[string]trace.Trace
	Links   map[string][]trace.Trace
	Number  map[string]float64
	Numbers map[string][]float64
	Str     map[string]string
	Strs    map[string][]string
	Time    map[string]time.Time
	Times   map[string][]time.Time

	BoolsSeen   map[string]map[bool]struct{}
	EnumsSeen   map[string]map[xopconst.Enum]struct{}
	Int64sSeen  map[string]map[int64]struct{}
	NumbersSeen map[string]map[float64]struct{}
	StrsSeen    map[string]map[string]struct{}

	LinksSeen map[string]map[string]struct{}
	TimesSeen map[string]map[int64]struct{}
}
