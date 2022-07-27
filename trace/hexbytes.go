// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package trace

import (
	"encoding/hex"
	"strings"
)

type HexBytes1 struct {
	b [1]byte
	s string
}
type HexBytes16 struct {
	b [16]byte
	s string
}
type HexBytes8 struct {
	b [8]byte
	s string
}

func NewHexBytes1() HexBytes1 {
	return HexBytes1{
		s: strings.Repeat("0", 1*2),
	}
}

func NewHexBytes16() HexBytes16 {
	return HexBytes16{
		s: strings.Repeat("0", 16*2),
	}
}

func NewHexBytes8() HexBytes8 {
	return HexBytes8{
		s: strings.Repeat("0", 8*2),
	}
}

var (
	zeroHexBytes1  = HexBytes1{}
	zeroHexBytes16 = HexBytes16{}
	zeroHexBytes8  = HexBytes8{}
)

func (x HexBytes1) IsZero() bool   { return x.b == zeroHexBytes1.b }
func (x HexBytes1) String() string { return x.s }
func (x HexBytes1) Bytes() []byte  { return x.b[:] }
func (x *HexBytes1) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	x.s = hex.EncodeToString(x.b[:])
}
func (x HexBytes16) IsZero() bool   { return x.b == zeroHexBytes16.b }
func (x HexBytes16) String() string { return x.s }
func (x HexBytes16) Bytes() []byte  { return x.b[:] }
func (x *HexBytes16) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	x.s = hex.EncodeToString(x.b[:])
}
func (x HexBytes8) IsZero() bool   { return x.b == zeroHexBytes8.b }
func (x HexBytes8) String() string { return x.s }
func (x HexBytes8) Bytes() []byte  { return x.b[:] }
func (x *HexBytes8) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes1) SetString(s string) {
	setBytesFromString(x.b[:], s)
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes16) SetString(s string) {
	setBytesFromString(x.b[:], s)
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes8) SetString(s string) {
	setBytesFromString(x.b[:], s)
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes1) SetZero() {
	x.b = [1]byte{}
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes16) SetZero() {
	x.b = [16]byte{}
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes8) SetZero() {
	x.b = [8]byte{}
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes1) SetRandom() {
	randomBytesNotAllZero(x.b[:])
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes16) SetRandom() {
	randomBytesNotAllZero(x.b[:])
	x.s = hex.EncodeToString(x.b[:])
}

func (x *HexBytes8) SetRandom() {
	randomBytesNotAllZero(x.b[:])
	x.s = hex.EncodeToString(x.b[:])
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
