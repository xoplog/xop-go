package xm

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field

func copyFields(from []Field) []Field {
	n := make([]Field, len(from))
	copy(n, from)
	return n
}

// TODO: how are we going to handle redacting Any?

func Namespace(key string) Field               { return zap.Namespace(key) }
func Skip() Field                              { return zap.Skip() }
func Stack(key string) Field                   { return zap.Stack(key) }
func Inline(val zapcore.ObjectMarshaler) Field { return zap.Inline(val) }
func Error(err error) Field                    { return zap.Error(err) }

// The following lines machine derived from the zap documentation page
// and thus potentially copyright Uber
// perl -p -e 's/^(func (\w+).*?, (\w+) .*)/\1 { return zap.\2(key, \3) }/'

func Any(key string, val interface{}) Field                { return zap.Any(key, val) }
func Array(key string, val zapcore.ArrayMarshaler) Field   { return zap.Array(key, val) }
func Binary(key string, val []byte) Field                  { return zap.Binary(key, val) }
func Bool(key string, val bool) Field                      { return zap.Bool(key, val) }
func Boolp(key string, val *bool) Field                    { return zap.Boolp(key, val) }
func Bools(key string, bs []bool) Field                    { return zap.Bools(key, bs) }
func ByteString(key string, val []byte) Field              { return zap.ByteString(key, val) }
func ByteStrings(key string, bss [][]byte) Field           { return zap.ByteStrings(key, bss) }
func Complex128(key string, val complex128) Field          { return zap.Complex128(key, val) }
func Complex128p(key string, val *complex128) Field        { return zap.Complex128p(key, val) }
func Complex128s(key string, nums []complex128) Field      { return zap.Complex128s(key, nums) }
func Complex64(key string, val complex64) Field            { return zap.Complex64(key, val) }
func Complex64p(key string, val *complex64) Field          { return zap.Complex64p(key, val) }
func Complex64s(key string, nums []complex64) Field        { return zap.Complex64s(key, nums) }
func Duration(key string, val time.Duration) Field         { return zap.Duration(key, val) }
func Durationp(key string, val *time.Duration) Field       { return zap.Durationp(key, val) }
func Durations(key string, ds []time.Duration) Field       { return zap.Durations(key, ds) }
func Errors(key string, errs []error) Field                { return zap.Errors(key, errs) }
func Float32(key string, val float32) Field                { return zap.Float32(key, val) }
func Float32p(key string, val *float32) Field              { return zap.Float32p(key, val) }
func Float32s(key string, nums []float32) Field            { return zap.Float32s(key, nums) }
func Float64(key string, val float64) Field                { return zap.Float64(key, val) }
func Float64p(key string, val *float64) Field              { return zap.Float64p(key, val) }
func Float64s(key string, nums []float64) Field            { return zap.Float64s(key, nums) }
func Int(key string, val int) Field                        { return zap.Int(key, val) }
func Int16(key string, val int16) Field                    { return zap.Int16(key, val) }
func Int16p(key string, val *int16) Field                  { return zap.Int16p(key, val) }
func Int16s(key string, nums []int16) Field                { return zap.Int16s(key, nums) }
func Int32(key string, val int32) Field                    { return zap.Int32(key, val) }
func Int32p(key string, val *int32) Field                  { return zap.Int32p(key, val) }
func Int32s(key string, nums []int32) Field                { return zap.Int32s(key, nums) }
func Int64(key string, val int64) Field                    { return zap.Int64(key, val) }
func Int64p(key string, val *int64) Field                  { return zap.Int64p(key, val) }
func Int64s(key string, nums []int64) Field                { return zap.Int64s(key, nums) }
func Int8(key string, val int8) Field                      { return zap.Int8(key, val) }
func Int8p(key string, val *int8) Field                    { return zap.Int8p(key, val) }
func Int8s(key string, nums []int8) Field                  { return zap.Int8s(key, nums) }
func Intp(key string, val *int) Field                      { return zap.Intp(key, val) }
func Ints(key string, nums []int) Field                    { return zap.Ints(key, nums) }
func NamedError(key string, err error) Field               { return zap.NamedError(key, err) }
func Object(key string, val zapcore.ObjectMarshaler) Field { return zap.Object(key, val) }
func Reflect(key string, val interface{}) Field            { return zap.Reflect(key, val) }
func StackSkip(key string, skip int) Field                 { return zap.StackSkip(key, skip) }
func String(key string, val string) Field                  { return zap.String(key, val) }
func Stringer(key string, val fmt.Stringer) Field          { return zap.Stringer(key, val) }
func Stringp(key string, val *string) Field                { return zap.Stringp(key, val) }
func Strings(key string, ss []string) Field                { return zap.Strings(key, ss) }
func Time(key string, val time.Time) Field                 { return zap.Time(key, val) }
func Timep(key string, val *time.Time) Field               { return zap.Timep(key, val) }
func Times(key string, ts []time.Time) Field               { return zap.Times(key, ts) }
func Uint(key string, val uint) Field                      { return zap.Uint(key, val) }
func Uint16(key string, val uint16) Field                  { return zap.Uint16(key, val) }
func Uint16p(key string, val *uint16) Field                { return zap.Uint16p(key, val) }
func Uint16s(key string, nums []uint16) Field              { return zap.Uint16s(key, nums) }
func Uint32(key string, val uint32) Field                  { return zap.Uint32(key, val) }
func Uint32p(key string, val *uint32) Field                { return zap.Uint32p(key, val) }
func Uint32s(key string, nums []uint32) Field              { return zap.Uint32s(key, nums) }
func Uint64(key string, val uint64) Field                  { return zap.Uint64(key, val) }
func Uint64p(key string, val *uint64) Field                { return zap.Uint64p(key, val) }
func Uint64s(key string, nums []uint64) Field              { return zap.Uint64s(key, nums) }
func Uint8(key string, val uint8) Field                    { return zap.Uint8(key, val) }
func Uint8p(key string, val *uint8) Field                  { return zap.Uint8p(key, val) }
func Uint8s(key string, nums []uint8) Field                { return zap.Uint8s(key, nums) }
func Uintp(key string, val *uint) Field                    { return zap.Uintp(key, val) }
func Uintptr(key string, val uintptr) Field                { return zap.Uintptr(key, val) }
func Uintptrp(key string, val *uintptr) Field              { return zap.Uintptrp(key, val) }
func Uintptrs(key string, us []uintptr) Field              { return zap.Uintptrs(key, us) }
func Uints(key string, nums []uint) Field                  { return zap.Uints(key, nums) }
