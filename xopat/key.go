package xopat

import (
	"github.com/xoplog/xop-go/xoputil"

	"github.com/muir/gwrap"
)

type K string

func (k K) String() string { return string(k) }

// JSONBody is the original key with escapes needed for being used
// as a JSON string body. Do not modify the slice.
func (k K) JSONBody() []byte {
	qjs := k.JSON()
	return qjs[1 : len(qjs)-1]
}

// JSONBody is the original key as a quoted JSON string.
// Do not modify the slice.
func (k K) JSON() []byte {
	qjs, ok := cachedKeys.Load(k)
	if ok {
		return qjs
	}
	b := xoputil.JBuilder{}
	b.B = make([]byte, 0, len(k)+2)
	b.AppendByte('"')
	b.AddStringBody(string(k))
	b.AppendByte('"')
	qjs, _ = cachedKeys.LoadOrStore(k, b.B)
	return qjs
}

var cachedKeys gwrap.SyncMap[K, []byte]

func ResetCachedKeys() {
	cachedKeys.Range(func(key K, value []byte) bool {
		cachedKeys.Delete(key)
		return true
	})
}
