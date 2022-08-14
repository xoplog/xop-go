package xoputil

type Prealloc struct {
	b []byte
}

func NewPrealloc(n []byte) *Prealloc {
	var p Prealloc
	p.Set(n)
	return &p
}

func (p *Prealloc) Set(n []byte) {
	p.b = n
}

func (p *Prealloc) Pack(n []byte) []byte {
	if len(n) > len(p.b) {
		return n
	}
	c := p.b[:len(n)]
	p.b = p.b[len(n):]
	copy(c, n)
	return c
}
