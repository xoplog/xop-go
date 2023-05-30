// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xopconsole

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoptrace"
	"github.com/xoplog/xop-go/xoputil"
)

const (
	numSinglePrealloc = 20
	numMultiPrealloc  = 12
)

// Writer implements io.Writer interface for json.Encoder
func (a *AttributeBuilder) Write(n []byte) (int, error) {
	*a.encodeTarget = append(*a.encodeTarget, n...)
	return len(n), nil
}

type AttributeBuilder struct {
	lock         sync.Mutex
	singlesBuf   [numSinglePrealloc]singleAttribute
	multiBuf     [numMultiPrealloc]multiAttribute
	singles      []singleAttribute
	multis       []multiAttribute
	Type         xopbase.DataType
	singleMap    map[string]*singleAttribute
	multiMap     map[string]*multiAttribute
	anyChanged   bool
	encodeTarget *[]byte
	encoder      *json.Encoder
}

type singleAttribute struct {
	attribute
	KeyValue []byte
	Buf      [40]byte
	keyLen   int
}

type multiAttribute struct {
	attribute
	Buf      [100]byte
	Distinct map[string]struct{}
	Builder  Builder
}

type attribute struct {
	Changed bool
	Type    xopbase.DataType
}

func (a *AttributeBuilder) Init() {
	a.singles = a.singlesBuf[:0]
	a.multis = a.multiBuf[:0]
	a.singleMap = make(map[string]*singleAttribute)
	a.multiMap = make(map[string]*multiAttribute)
	a.anyChanged = false
}

// Append will only add data if there is any unflushed data to add.
func (a *AttributeBuilder) Append(b *Builder, onlyChanged bool, attributesObject bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if (!a.anyChanged && onlyChanged) || (len(a.multiMap) == 0 && len(a.singleMap) == 0) {
		return
	}
	a.anyChanged = false
	for _, m := range a.multiMap {
		if m.Changed || !onlyChanged {
			b.AppendByte(' ')
			b.AppendBytes(m.Builder.B)
			m.Changed = false
		}
	}
	for _, s := range a.singleMap {
		if s.Changed || !onlyChanged {
			b.AppendByte(' ')
			b.AppendBytes(s.KeyValue)
			s.Changed = false
		}
	}
}

func (m *multiAttribute) init(a *AttributeBuilder, k xopat.AttributeInterface) {
	m.Builder.B = m.Buf[:0]
	m.Builder.Reset()
	m.Builder.AppendBytes(k.ConsoleKey())
	m.Builder.AppendByte('=')
	m.Distinct = nil
}

func (a *AttributeBuilder) addMulti(k xopat.AttributeInterface) *multiAttribute {
	var m *multiAttribute
	var ok bool
	m, ok = a.multiMap[k.Key()]
	if !ok {
		if len(a.multis) == cap(a.multis) {
			a.multis = make([]multiAttribute, 0, cap(a.multis))
		}
		a.multis = a.multis[:len(a.multis)+1]
		m = &a.multis[len(a.multis)-1]
		m.init(a, k)
		a.multiMap[k.Key()] = m
	}
	m.Changed = true
	return m
}

func (s *singleAttribute) init(k xopat.AttributeInterface) {
	b := xoputil.JBuilder{
		B: s.Buf[:0],
	}
	b.AppendBytes(k.ConsoleKey())
	s.Changed = true
	s.KeyValue = b.B
}

func (a *AttributeBuilder) addSingle(k xopat.AttributeInterface) (*singleAttribute, bool) {
	s, ok := a.singleMap[k.Key()]
	if !ok {
		if len(a.singles) == cap(a.singles) {
			a.singles = make([]singleAttribute, 0, cap(a.singles))
		}
		a.singles = a.singles[:len(a.singles)+1]
		s = &a.singles[len(a.singles)-1]
		s.init(k)
		a.singleMap[k.Key()] = s
	}
	s.Changed = true
	return s, ok
}

func (a *AttributeBuilder) MetadataAny(k *xopat.AnyAttribute, v xopbase.ModelArg) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if a.encoder == nil {
		a.encoder = json.NewEncoder(a)
		a.encoder.SetEscapeHTML(false)
	}
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.AnyDataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
			encoder: a.encoder,
		}
		a.encodeTarget = &b.B
		b.AttributeAny(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.AnyDataType
	a.encodeTarget = &m.Builder.B
	m.Builder.encoder = a.encoder
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeAny(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}

func (a *AttributeBuilder) MetadataBool(k *xopat.BoolAttribute, v bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.BoolDataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AttributeBool(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.BoolDataType
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeBool(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}

func (a *AttributeBuilder) MetadataEnum(k *xopat.EnumAttribute, v xopat.Enum) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.EnumDataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AttributeEnum(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.EnumDataType
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeEnum(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}

func (a *AttributeBuilder) MetadataFloat64(k *xopat.Float64Attribute, v float64) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.Float64DataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AttributeFloat64(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.Float64DataType
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeFloat64(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}

func (a *AttributeBuilder) MetadataInt64(k *xopat.Int64Attribute, v int64) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.Int64DataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AttributeInt64(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.Int64DataType
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeInt64(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}

func (a *AttributeBuilder) MetadataLink(k *xopat.LinkAttribute, v xoptrace.Trace) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.LinkDataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AttributeLink(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.LinkDataType
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeLink(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}

func (a *AttributeBuilder) MetadataString(k *xopat.StringAttribute, v string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.StringDataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AttributeString(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.StringDataType
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeString(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}

func (a *AttributeBuilder) MetadataTime(k *xopat.TimeAttribute, v time.Time) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.anyChanged = true
	if !k.Multiple() {
		s, preExisting := a.addSingle(k)
		if preExisting {
			if k.Locked() {
				return
			} else {
				s.KeyValue = s.KeyValue[:s.keyLen]
			}
		} else {
			s.keyLen = len(s.KeyValue)
		}
		s.Type = xopbase.TimeDataType
		b := Builder{
			JBuilder: xoputil.JBuilder{
				B: s.KeyValue,
			},
		}
		b.AttributeTime(v)
		s.KeyValue = b.B
		return
	}
	m := a.addMulti(k)
	m.Type = xopbase.TimeDataType
	lenBeforeKey := len(m.Builder.B)
	if len(m.Builder.B) != 0 {
		m.Builder.AppendByte(' ')
	}
	// we add the new value unconditionally but can retroactively remove it by shortening to lenBeforeKey
	m.Builder.AppendBytes(k.ConsoleKey())
	lenBeforeData := len(m.Builder.B)
	m.Builder.AttributeTime(v)
	if k.Distinct() {
		sk := string(m.Builder.B[lenBeforeData:len(m.Builder.B)])
		if m.Distinct == nil {
			m.Distinct = make(map[string]struct{})
			m.Distinct[sk] = struct{}{}
		} else {
			if _, ok := m.Distinct[sk]; ok {
				m.Builder.B = m.Builder.B[:lenBeforeKey]
				if m.Builder.B[len(m.Builder.B)-1] == ',' {
					m.Builder.B = m.Builder.B[0 : len(m.Builder.B)-1]
				}
			} else {
				m.Distinct[sk] = struct{}{}
			}
		}
	}
}
