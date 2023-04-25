package xopbaseutil

import (
	"encoding/json"

	"github.com/muir/gwrap"

	"github.com/xoplog/xop-go/xopbase"
	"github.com/xoplog/xop-go/xoputil"
)

var _ xopbase.Builder = &Builder{}

type Builder struct {
	xoputil.JBuilder
	encoder *json.Encoder
}

func (b *Builder) Initialize() {
	b.encoder = json.NewEncoder(&b.JBuilder)
	b.encoder.SetEscapeHTML(false)
}

func (b *Builder) Reset(s *span) {
	b.B = b.B[:0]
}

type containsBuilder[W containsBuilder] interface {
	Reset() // must call BaseBuilder().Reset()
	BaseBuilder() *Builder
}

// BuilderPool is a pool of builders. When creating
// a BuilderPool, Pool.New must be a function that
// returns a W and calls Builder.Initialize().
type BuilderPool[W containsBuilder] struct {
	Pool      gwrap.SyncPool[W]
	MinBuffer int
	MaxBuffer int
}

func (bp BuilderPool[W]) Get() W {
	n := bp.Pool.Get()
	return n
}

func (bp BuilderPool[W]) Reclaim(w W) {
	if len(w.BaseBuilder().B) > bp.MaxBuffer {
		// we're not putting this one back in the pool
		// because it's too big and it's better to just
		// free it's memory.
		return
	}
	w.Reset()
	bp.Pool.Put(w)
}
