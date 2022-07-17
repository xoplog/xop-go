// This file is generated, DO NOT EDIT
// It is generated from the corresponding .zzzgo file using zopzzz
//
package xoputil

import (
	"time"

	"github.com/muir/xoplog/trace"
	"github.com/muir/xoplog/xopconst"
)

type AttributeBuilder struct {
	Any           map[string]interface{}
	Anys          map[string][]interface{}
	AnysSeen      map[string]map[interface{}]struct{}
	Bool          map[string]bool
	Bools         map[string][]bool
	BoolsSeen     map[string]map[bool]struct{}
	Duration      map[string]time.Duration
	Durations     map[string][]time.Duration
	DurationsSeen map[string]map[time.Duration]struct{}
	Int           map[string]int
	Ints          map[string][]int
	IntsSeen      map[string]map[int]struct{}
	Link          map[string]trace.Trace
	Links         map[string][]trace.Trace
	LinksSeen     map[string]map[trace.Trace]struct{}
	Str           map[string]string
	Strs          map[string][]string
	StrsSeen      map[string]map[string]struct{}
	Time          map[string]time.Time
	Times         map[string][]time.Time
	TimesSeen     map[string]map[time.Time]struct{}
}

func NewAttributeBuilder() *AttributeBuilder {
	return &AttributeBuilder{
		Any:           make(map[string]interface{}),
		Anys:          make(map[string][]interface{}),
		AnysSeen:      make(map[string]map[interface{}]struct{}),
		Bool:          make(map[string]bool),
		Bools:         make(map[string][]bool),
		BoolsSeen:     make(map[string]map[bool]struct{}),
		Duration:      make(map[string]time.Duration),
		Durations:     make(map[string][]time.Duration),
		DurationsSeen: make(map[string]map[time.Duration]struct{}),
		Int:           make(map[string]int),
		Ints:          make(map[string][]int),
		IntsSeen:      make(map[string]map[int]struct{}),
		Link:          make(map[string]trace.Trace),
		Links:         make(map[string][]trace.Trace),
		LinksSeen:     make(map[string]map[trace.Trace]struct{}),
		Str:           make(map[string]string),
		Strs:          make(map[string][]string),
		StrsSeen:      make(map[string]map[string]struct{}),
		Time:          make(map[string]time.Time),
		Times:         make(map[string][]time.Time),
		TimesSeen:     make(map[string]map[time.Time]struct{}),
	}
}

func (a *AttributeBuilder) MetadataAny(k *xopconst.AnyAttribute, v interface{}) {
	if k.Multiple() {
		ary = append(ary, v)
		a.Any[k.Key()] = append(a.Any[k.Key()], v)
	} else {
		a.Any[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
	if k.Multiple() {
		ary = append(ary, v)
		if k.Distinct() {
			if seenMap, ok := a.TimesSeen[k.Key()]; ok {
				if _, ok := seenMap[v.UnixNano()]; ok {
					return
				}
			} else {
				z.TimesSeen[k.Key()] = make(map[int64]struct{})
			}
			z.TimesSeen[k.Key()][v.UnixNano()] = struct{}{}
		}
		a.Time[k.Key()] = append(a.Time[k.Key()], v)
	} else {
		a.Time[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataBool(k *xopconst.BoolAttribute, v bool) {
	if k.Multiple() {
		ary = append(ary, v)
		if k.Distinct() {
			if seenMap, ok := a.BoolsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				z.BoolsSeen[k.Key()] = make(map[bool]struct{})
			}
			z.BoolsSeen[k.Key()][v] = struct{}{}
		}
		a.Bool[k.Key()] = append(a.Bool[k.Key()], v)
	} else {
		a.Bool[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataDuration(k *xopconst.DurationAttribute, v time.Duration) {
	if k.Multiple() {
		ary = append(ary, v)
		if k.Distinct() {
			if seenMap, ok := a.DurationsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				z.DurationsSeen[k.Key()] = make(map[time.Duration]struct{})
			}
			z.DurationsSeen[k.Key()][v] = struct{}{}
		}
		a.Duration[k.Key()] = append(a.Duration[k.Key()], v)
	} else {
		a.Duration[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataInt(k *xopconst.IntAttribute, v int) {
	if k.Multiple() {
		ary = append(ary, v)
		if k.Distinct() {
			if seenMap, ok := a.IntsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				z.IntsSeen[k.Key()] = make(map[int]struct{})
			}
			z.IntsSeen[k.Key()][v] = struct{}{}
		}
		a.Int[k.Key()] = append(a.Int[k.Key()], v)
	} else {
		a.Int[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	if k.Multiple() {
		ary = append(ary, v)
		if k.Distinct() {
			if seenMap, ok := a.LinksSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				z.LinksSeen[k.Key()] = make(map[trace.Trace]struct{})
			}
			z.LinksSeen[k.Key()][v] = struct{}{}
		}
		a.Link[k.Key()] = append(a.Link[k.Key()], v)
	} else {
		a.Link[k.Key()] = v
	}
}

func (a *AttributeBuilder) MetadataStr(k *xopconst.StrAttribute, v string) {
	if k.Multiple() {
		ary = append(ary, v)
		if k.Distinct() {
			if seenMap, ok := a.StrsSeen[k.Key()]; ok {
				if _, ok := seenMap[v]; ok {
					return
				}
			} else {
				z.StrsSeen[k.Key()] = make(map[string]struct{})
			}
			z.StrsSeen[k.Key()][v] = struct{}{}
		}
		a.Str[k.Key()] = append(a.Str[k.Key()], v)
	} else {
		a.Str[k.Key()] = v
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
	if len(a.Duration) != 0 {
		for k, v := range a.Duration {
			m[k] = v
		}
	}
	if len(a.Durations) != 0 {
		for k, v := range a.Durations {
			m[k] = v
		}
	}
	if len(a.Int) != 0 {
		for k, v := range a.Int {
			m[k] = v
		}
	}
	if len(a.Ints) != 0 {
		for k, v := range a.Ints {
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
