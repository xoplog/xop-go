package xopat

import (
	"github.com/xoplog/xop-go/xoputil"
)

var defineAttributeStarter = []byte(`{"type":"defineKey","key":"`) // }

func jsonAttributeDefinition(k *Attribute) []byte {
	b := xoputil.JBuilder{}
	b.B = make([]byte, len(defineAttributeStarter), len(defineAttributeStarter)+100)
	copy(b.B, defineAttributeStarter)
	b.AddStringBody(k.Key())
	b.AppendBytes([]byte(`","desc":"`))
	b.AddStringBody(k.Description())
	b.AppendBytes([]byte(`","ns":"`))
	b.AddStringBody(k.Namespace())
	b.AppendByte(' ')
	b.AddStringBody(k.SemverString())
	if k.Indexed() {
		b.AppendBytes([]byte(`","indexed":true,"prom":`))
	} else {
		b.AppendBytes([]byte(`","prom":`))
	}
	b.AddInt32(int32(k.Prominence()))
	if k.Multiple() {
		if k.Distinct() {
			b.AppendBytes([]byte(`,"mult":true,"distinct":true,"vtype":`))
		} else {
			b.AppendBytes([]byte(`,"mult":true,"vtype":`))
		}
	} else {
		if k.Locked() {
			b.AppendBytes([]byte(`,"locked":true,"vtype":`))
		} else {
			b.AppendBytes([]byte(`,"vtype":`))
		}
	}
	b.AddSafeString(k.ProtoType().String())
	if k.Ranged() {
		b.AppendBytes([]byte(`,"ranged":true`))
	}
	// {
	b.AppendBytes([]byte("}\n"))
	return b.B
}
