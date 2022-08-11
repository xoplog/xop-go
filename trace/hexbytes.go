// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package trace

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
)

type HexBytes1 struct {
	b [1]byte
	h [1 * 2]byte
}
type HexBytes16 struct {
	b [16]byte
	h [16 * 2]byte
}
type HexBytes8 struct {
	b [8]byte
	h [8 * 2]byte
}

var (
	zeroHexHexBytes1  = bytes.Repeat([]byte{'0'}, 1*2)
	zeroHexHexBytes16 = bytes.Repeat([]byte{'0'}, 16*2)
	zeroHexHexBytes8  = bytes.Repeat([]byte{'0'}, 8*2)
)

func NewHexBytes1() HexBytes1 {
	var x HexBytes1
	copy(x.h[:], zeroHexHexBytes1)
	return x
}

func NewHexBytes16() HexBytes16 {
	var x HexBytes16
	copy(x.h[:], zeroHexHexBytes16)
	return x
}

func NewHexBytes8() HexBytes8 {
	var x HexBytes8
	copy(x.h[:], zeroHexHexBytes8)
	return x
}

var (
	zeroHexBytes1b  = [1]byte{}
	zeroHexBytes16b = [16]byte{}
	zeroHexBytes8b  = [8]byte{}
)

func (x HexBytes1) IsZero() bool      { return x.b == zeroHexBytes1b }
func (x HexBytes1) String() string    { return string(x.h[:]) }
func (x HexBytes1) Bytes() []byte     { return x.b[:] }
func (x HexBytes1) HexBytes() []byte  { return x.h[:] }
func (x HexBytes16) IsZero() bool     { return x.b == zeroHexBytes16b }
func (x HexBytes16) String() string   { return string(x.h[:]) }
func (x HexBytes16) Bytes() []byte    { return x.b[:] }
func (x HexBytes16) HexBytes() []byte { return x.h[:] }
func (x HexBytes8) IsZero() bool      { return x.b == zeroHexBytes8b }
func (x HexBytes8) String() string    { return string(x.h[:]) }
func (x HexBytes8) Bytes() []byte     { return x.b[:] }
func (x HexBytes8) HexBytes() []byte  { return x.h[:] }

func (x *HexBytes1) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes16) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes8) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes1) SetString(s string) {
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes16) SetString(s string) {
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes8) SetString(s string) {
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes1) SetZero() {
	x.b = [1]byte{}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes16) SetZero() {
	x.b = [16]byte{}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes8) SetZero() {
	x.b = [8]byte{}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes1) SetRandom() {
	for {
		_, _ = rand.Read(x.b[:])
		if !allZero(x.b[:]) {
			break
		}
	}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes16) SetRandom() {
	for {
		_, _ = rand.Read(x.b[:])
		if !allZero(x.b[:]) {
			break
		}
	}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes8) SetRandom() {
	for {
		_, _ = rand.Read(x.b[:])
		if !allZero(x.b[:]) {
			break
		}
	}
	hex.Encode(x.h[:], x.b[:])
}

func (x HexBytes1) Copy() HexBytes1 {
	return x
}

func (x HexBytes16) Copy() HexBytes16 {
	return x
}

func (x HexBytes8) Copy() HexBytes8 {
	return x
}
