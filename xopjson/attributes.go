// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopjson

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/muir/xop-go/trace"
	"github.com/muir/xop-go/xopconst"
	"github.com/muir/xop-go/xoputil"
)

const (
	minMap            = 12
	numSinglePrealloc = 20
	numMultiPrealloc  = 12
)

// Writer implements io.Writer interface for json.Encoder
func (a AttributeBuilder) Write(n []byte) (int, error) {
	*a.encodeTarget = append(*a.encodeTarget, n...)
	return len(n), nil
}

type AttributeBuilder struct {
	lock         sync.Mutex
	singlesBuf   [numSinglePrealloc]singleAttribute
	multiBuf     [numMultiPrealloc]multiAttribute
	singles      []singleAttribute
	multis       []multiAttribute
	Type         xoputil.BaseAttributeType
	singleMap    map[string]*singleAttribute // available only when count > minMap
	multiMap     map[string]*multiAttribute  // available only when count > minMap
	anyChanged   bool
	span         *span
	encodeTarget *[]byte
	encoder      *json.Encoder
}

type singleAttribute struct {
	attribute
	KeyValue []byte
	Buf      [40]byte
}

type multiAttribute struct {
	attribute
	Buf      [100]byte
	Distinct map[string]struct{}
	Builder  builder
}

type attribute struct {
	Name    []byte
	Changed bool
	Type    xoputil.BaseAttributeType
}

func (a *AttributeBuilder) Init(s *span) {
	a.singles = a.singlesBuf[:0]
	a.multis = a.multiBuf[:0]
	a.singleMap = nil
	a.multiMap = nil
	a.anyChanged = false
	a.span = s
}

func (a *AttributeBuilder) Append(b *xoputil.JBuilder) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if !a.anyChanged {
		return
	}
	a.anyChanged = false
	multi := func(m *multiAttribute) {
		if m.Changed {
			b.Comma()
			b.AppendBytes(m.Builder.B)
			// [
			b.AppendByte(']')
			m.Changed = false
		}
	}
	if a.multiMap != nil {
		for _, m := range a.multiMap {
			multi(m)
		}
	} else {
		for i := range a.multis {
			multi(&a.multis[i])
		}
	}
	single := func(s *singleAttribute) {
		if s.Changed {
			b.Comma()
			b.AppendBytes(s.KeyValue)
			s.Changed = false
		}
	}
	if a.singleMap != nil {
		for _, s := range a.singleMap {
			single(s)
		}
	} else {
		for i := range a.singles {
			single(&a.singles[i])
		}
	}
}

func (a *AttributeBuilder) useMultiMap() {
	a.multiMap = make(map[string]*multiAttribute)
	for i, p := range a.multis {
		a.multiMap[string(p.Name)] = &a.multis[i]
	}
}

func (m *multiAttribute) init(a *AttributeBuilder, k string) {
	m.Builder.B = m.Buf[:0]
	m.Builder.reset(a.span)
	m.Builder.AddString(k)
	if len(m.Builder.B) == len(k)+2 {
		m.Name = m.Builder.B[1 : len(m.Builder.B)-1]
	} else {
		m.Name = []byte(k)
	}
	m.Builder.AppendBytes([]byte{':', '['}) // ]
	m.Distinct = nil
}

func (a *AttributeBuilder) addMulti(k string) *multiAttribute {
	var m *multiAttribute
	var ok bool
	if a.multiMap != nil {
		m, ok = a.multiMap[k]
	} else {
		for _, m := range a.multis {
			if k == string(m.Name) {
				ok = true
				break
			}
		}
	}
	if !ok {
		if cap(a.multis) > len(a.multis) {
			a.multis = a.multis[:len(a.multis)+1]
			m = &a.multis[len(a.multis)-1]
		} else {
			m = &multiAttribute{}
			a.useMultiMap()
		}
		m.init(a, k)
		if a.multiMap != nil {
			a.multiMap[k] = m
		}
	}
	m.Changed = true
	m.Builder.Comma()
	return m
}

func (a *AttributeBuilder) useSingleMap() {
	a.singleMap = make(map[string]*singleAttribute)
	for i, p := range a.singles {
		a.singleMap[string(p.Name)] = &a.singles[i]
	}
}

func (s *singleAttribute) init(k string) {
	b := xoputil.JBuilder{
		B: s.Buf[:0],
	}
	b.AddString(k)
	if len(b.B) == len(k)+2 {
		s.Name = b.B[1 : len(b.B)-1]
	} else {
		s.Name = []byte(k)
	}
	b.AppendByte(':')
	s.Changed = false
	s.KeyValue = b.B
}

func (a *AttributeBuilder) addSingle(k string) *singleAttribute {
	var s *singleAttribute
	var ok bool
	if a.singleMap != nil {
		s, ok = a.singleMap[k]
	} else {
		for _, s := range a.singles {
			if k == string(s.Name) {
				ok = true
				break
			}
		}
	}
	if !ok {
		if cap(a.singles) > len(a.singles) {
			a.singles = a.singles[:len(a.singles)+1]
			s = &a.singles[len(a.singles)-1]
		} else {
			s = &singleAttribute{}
			a.useSingleMap()
		}
		s.init(k)
		if a.singleMap != nil {
			a.singleMap[k] = s
		}
	}
	s.Changed = true
	return s
}

func (a *AttributeBuilder) MetadataAny(k *xopconst.AnyAttribute, v interface{}) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.encoder == nil {
		a.encoder = json.NewEncoder(a)
		a.encoder.SetEscapeHTML(false)
	}
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeAnyArray
		m.Builder.Comma()
		a.encodeTarget = &m.Builder.B
		m.Builder.encoder = a.encoder
		lenBefore := len(m.Builder.B)
		m.Builder.AddAny(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeAny
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
			encoder: a.encoder,
		}
		a.encodeTarget = &b.B
		b.AddAny(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}

func (a *AttributeBuilder) MetadataBool(k *xopconst.BoolAttribute, v bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeBoolArray
		m.Builder.Comma()
		lenBefore := len(m.Builder.B)
		m.Builder.AddBool(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeBool
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AddBool(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}

func (a *AttributeBuilder) MetadataEnum(k *xopconst.EnumAttribute, v xopconst.Enum) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeEnumArray
		m.Builder.Comma()
		lenBefore := len(m.Builder.B)
		m.Builder.AddEnum(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeEnum
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AddEnum(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}

func (a *AttributeBuilder) MetadataFloat64(k *xopconst.Float64Attribute, v float64) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeFloat64Array
		m.Builder.Comma()
		lenBefore := len(m.Builder.B)
		m.Builder.AddFloat64(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeFloat64
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AddFloat64(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}

func (a *AttributeBuilder) MetadataInt64(k *xopconst.Int64Attribute, v int64) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeInt64Array
		m.Builder.Comma()
		lenBefore := len(m.Builder.B)
		m.Builder.AddInt64(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeInt64
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AddInt64(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}

func (a *AttributeBuilder) MetadataLink(k *xopconst.LinkAttribute, v trace.Trace) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeLinkArray
		m.Builder.Comma()
		lenBefore := len(m.Builder.B)
		m.Builder.AddLink(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeLink
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AddLink(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}

func (a *AttributeBuilder) MetadataString(k *xopconst.StringAttribute, v string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeStringArray
		m.Builder.Comma()
		lenBefore := len(m.Builder.B)
		m.Builder.AddString(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeString
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AddString(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}

func (a *AttributeBuilder) MetadataTime(k *xopconst.TimeAttribute, v time.Time) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if k.Multiple() {
		m := a.addMulti(k.Key())
		m.Type = xoputil.BaseAttributeTypeTimeArray
		m.Builder.Comma()
		lenBefore := len(m.Builder.B)
		m.Builder.AddTime(v)
		if k.Distinct() {
			sk := string(m.Builder.B[lenBefore:len(m.Builder.B)])
			if m.Distinct == nil {
				m.Distinct = make(map[string]struct{})
				m.Distinct[sk] = struct{}{}
			} else {
				if _, ok := m.Distinct[sk]; ok {
					m.Builder.B = m.Builder.B[:lenBefore]
				} else {
					m.Distinct[sk] = struct{}{}
				}
			}
		}
	} else {
		s := a.addSingle(k.Key())
		s.Type = xoputil.BaseAttributeTypeTime
		b := builder{
			span: a.span,
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AddTime(v)
		s.KeyValue = b.B
	}
	a.anyChanged = true
}
