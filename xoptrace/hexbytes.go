// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoptrace

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

type WrappedHexBytes1 struct {
	*HexBytes1
	offset int
	trace  *Trace
}
type WrappedHexBytes16 struct {
	*HexBytes16
	offset int
	trace  *Trace
}
type WrappedHexBytes8 struct {
	*HexBytes8
	offset int
	trace  *Trace
}

var (
	zeroHexHexBytes1  = bytes.Repeat([]byte{'0'}, 1*2)
	zeroHexHexBytes16 = bytes.Repeat([]byte{'0'}, 16*2)
	zeroHexHexBytes8  = bytes.Repeat([]byte{'0'}, 8*2)
)

func newHexBytes1() HexBytes1 {
	var x HexBytes1
	copy(x.h[:], zeroHexHexBytes1)
	return x
}

func newHexBytes16() HexBytes16 {
	var x HexBytes16
	copy(x.h[:], zeroHexHexBytes16)
	return x
}

func newHexBytes8() HexBytes8 {
	var x HexBytes8
	copy(x.h[:], zeroHexHexBytes8)
	return x
}

func NewHexBytes1FromSlice(b []byte) HexBytes1 {
	var x HexBytes1
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
	return x
}

func NewHexBytes16FromSlice(b []byte) HexBytes16 {
	var x HexBytes16
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
	return x
}

func NewHexBytes8FromSlice(b []byte) HexBytes8 {
	var x HexBytes8
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
	return x
}

func NewHexBytes1FromString(s string) HexBytes1 {
	var x HexBytes1
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
	return x
}

func NewHexBytes16FromString(s string) HexBytes16 {
	var x HexBytes16
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
	return x
}

func NewHexBytes8FromString(s string) HexBytes8 {
	var x HexBytes8
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
	return x
}

var (
	zeroHexBytes1b  = [1]byte{}
	zeroHexBytes16b = [16]byte{}
	zeroHexBytes8b  = [8]byte{}
)

func (x HexBytes1) IsZero() bool      { return x.b == zeroHexBytes1b }
func (x HexBytes1) Bytes() []byte     { return x.b[:] }
func (x HexBytes1) HexBytes() []byte  { return x.h[:] }
func (x HexBytes1) String() string    { return string(x.h[:]) }
func (x HexBytes16) IsZero() bool     { return x.b == zeroHexBytes16b }
func (x HexBytes16) Bytes() []byte    { return x.b[:] }
func (x HexBytes16) HexBytes() []byte { return x.h[:] }
func (x HexBytes16) String() string   { return string(x.h[:]) }
func (x HexBytes8) IsZero() bool      { return x.b == zeroHexBytes8b }
func (x HexBytes8) Bytes() []byte     { return x.b[:] }
func (x HexBytes8) HexBytes() []byte  { return x.h[:] }
func (x HexBytes8) String() string    { return string(x.h[:]) }

func (x WrappedHexBytes1) String() string {
	return x.trace.headerString[x.offset : x.offset+1*2]
}

func (x WrappedHexBytes16) String() string {
	return x.trace.headerString[x.offset : x.offset+16*2]
}

func (x WrappedHexBytes8) String() string {
	return x.trace.headerString[x.offset : x.offset+8*2]
}

func (x WrappedHexBytes1) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes16) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes8) SetBytes(b []byte) {
	setBytes(x.b[:], b)
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes1) SetArray(b [1]byte) {
	x.b = b
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes16) SetArray(b [16]byte) {
	x.b = b
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes8) SetArray(b [8]byte) {
	x.b = b
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

// Array returns the underlying byte array.  Do not modify it!
func (x *HexBytes1) Array() [1]byte {
	return x.b
}

// Array returns the underlying byte array.  Do not modify it!
func (x *HexBytes16) Array() [16]byte {
	return x.b
}

// Array returns the underlying byte array.  Do not modify it!
func (x *HexBytes8) Array() [8]byte {
	return x.b
}

func (x WrappedHexBytes1) SetString(s string) {
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes16) SetString(s string) {
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes8) SetString(s string) {
	setBytesFromString(x.b[:], s)
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes1) SetZero() {
	x.b = [1]byte{}
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes16) SetZero() {
	x.b = [16]byte{}
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes8) SetZero() {
	x.b = [8]byte{}
	hex.Encode(x.h[:], x.b[:])
	x.trace.rebuild()
}

func (x WrappedHexBytes1) SetRandom() {
	x.HexBytes1.setRandom()
	x.trace.rebuild()
}

func (x WrappedHexBytes16) SetRandom() {
	x.HexBytes16.setRandom()
	x.trace.rebuild()
}

func (x WrappedHexBytes8) SetRandom() {
	x.HexBytes8.setRandom()
	x.trace.rebuild()
}

func (x WrappedHexBytes1) Set(b HexBytes1) {
	*x.HexBytes1 = b
	x.trace.rebuild()
}

func (x WrappedHexBytes16) Set(b HexBytes16) {
	*x.HexBytes16 = b
	x.trace.rebuild()
}

func (x WrappedHexBytes8) Set(b HexBytes8) {
	*x.HexBytes8 = b
	x.trace.rebuild()
}

func (x *HexBytes1) setRandom() {
	for {
		_, _ = rand.Read(x.b[:])
		if x.b != zeroHexBytes1b {
			break
		}
	}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes16) setRandom() {
	for {
		_, _ = rand.Read(x.b[:])
		if x.b != zeroHexBytes16b {
			break
		}
	}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes8) setRandom() {
	for {
		_, _ = rand.Read(x.b[:])
		if x.b != zeroHexBytes8b {
			break
		}
	}
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes1) initialize() {
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes16) initialize() {
	hex.Encode(x.h[:], x.b[:])
}

func (x *HexBytes8) initialize() {
	hex.Encode(x.h[:], x.b[:])
}

func (x HexBytes1) initialized(t Trace) HexBytes1 {
	if !t.initialized {
		hex.Encode(x.h[:], x.b[:])
	}
	return x
}

func (x HexBytes16) initialized(t Trace) HexBytes16 {
	if !t.initialized {
		hex.Encode(x.h[:], x.b[:])
	}
	return x
}

func (x HexBytes8) initialized(t Trace) HexBytes8 {
	if !t.initialized {
		hex.Encode(x.h[:], x.b[:])
	}
	return x
}
