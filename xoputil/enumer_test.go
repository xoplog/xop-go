// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file

package xoputil_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xoplog/xop-go/xopnum"
	"github.com/xoplog/xop-go/xoprecorder"
)

func TestEnumerEventType(t *testing.T) {
	values := xoprecorder.EventTypeValues()
	vlast := values[len(values)-1]
	assert.NotEmpty(t, vlast.String(), "valid")
	assert.NotEmpty(t, (vlast + 1).String(), "invalid")
	for _, s := range xoprecorder.EventTypeStrings() {
		v, err := xoprecorder.EventTypeString(s)
		assert.NoError(t, err, s)
		assert.Equal(t, s, v.String())
	}
	v, err := xoprecorder.EventTypeString(vlast.String())
	require.NoError(t, err, "identity")
	assert.Equal(t, vlast, v, "identity")
	v, err = xoprecorder.EventTypeString(strings.ToLower(vlast.String()))
	require.NoError(t, err, "identity, lower")
	assert.Equal(t, vlast, v, "identity, lower")
	_, err = xoprecorder.EventTypeString("lasjf;asjfl;adsjf;lasdjfl;jasdf")
	assert.Error(t, err, "invalid")
	assert.True(t, vlast.IsAEventType(), "is valid")
	assert.False(t, (vlast + 1).IsAEventType(), "is not valid")
	enc, err := json.Marshal(vlast)
	require.NoError(t, err, "marshal")
	require.NotEmpty(t, enc, "enc")
	var unenc xoprecorder.EventType
	err = json.Unmarshal(enc, &unenc)
	require.NoError(t, err, "unmarshal")
	assert.Equal(t, vlast, unenc, "json round trip")
	err = json.Unmarshal([]byte(strconv.Itoa(int(vlast))), &unenc)
	assert.Error(t, err, "unmarshal error expected")
	value, err := vlast.Value()
	assert.NoError(t, err, "value")
	assert.NotEmpty(t, value, "value")
	var scan xoprecorder.EventType
	err = (&scan).Scan(vlast.String())
	assert.NoError(t, err, "scan string")
	assert.Equal(t, vlast, scan, "scan string")
	scan++
	err = (&scan).Scan([]byte(vlast.String()))
	assert.NoError(t, err, "scan bytes")
	assert.Equal(t, vlast, scan, "scan bytes")
	scan++
	err = (&scan).Scan(vlast)
	assert.NoError(t, err, "scan stringer")
	assert.Equal(t, vlast, scan, "scan stringer")
	scan++
	err = (&scan).Scan(int(vlast))
	assert.Error(t, err, "scan int")
	assert.NoError(t, (&scan).Scan(nil), "scan nil")
	assert.Error(t, err, (&scan).Scan("als;djfa;dlfja;ldfjal;sdjfl;sjdf;"), "invalid scan string")
}

func TestEnumerLevel(t *testing.T) {
	values := xopnum.LevelValues()
	vlast := values[len(values)-1]
	assert.NotEmpty(t, vlast.String(), "valid")
	assert.NotEmpty(t, (vlast + 1).String(), "invalid")
	for _, s := range xopnum.LevelStrings() {
		v, err := xopnum.LevelString(s)
		assert.NoError(t, err, s)
		assert.Equal(t, s, v.String())
	}
	v, err := xopnum.LevelString(vlast.String())
	require.NoError(t, err, "identity")
	assert.Equal(t, vlast, v, "identity")
	v, err = xopnum.LevelString(strings.ToLower(vlast.String()))
	require.NoError(t, err, "identity, lower")
	assert.Equal(t, vlast, v, "identity, lower")
	_, err = xopnum.LevelString("lasjf;asjfl;adsjf;lasdjfl;jasdf")
	assert.Error(t, err, "invalid")
	assert.True(t, vlast.IsALevel(), "is valid")
	assert.False(t, (vlast + 1).IsALevel(), "is not valid")
	enc, err := json.Marshal(vlast)
	require.NoError(t, err, "marshal")
	require.NotEmpty(t, enc, "enc")
	var unenc xopnum.Level
	err = json.Unmarshal(enc, &unenc)
	require.NoError(t, err, "unmarshal")
	assert.Equal(t, vlast, unenc, "json round trip")
	err = json.Unmarshal([]byte(strconv.Itoa(int(vlast))), &unenc)
	assert.Error(t, err, "unmarshal error expected")
	value, err := vlast.Value()
	assert.NoError(t, err, "value")
	assert.NotEmpty(t, value, "value")
	var scan xopnum.Level
	err = (&scan).Scan(vlast.String())
	assert.NoError(t, err, "scan string")
	assert.Equal(t, vlast, scan, "scan string")
	scan++
	err = (&scan).Scan([]byte(vlast.String()))
	assert.NoError(t, err, "scan bytes")
	assert.Equal(t, vlast, scan, "scan bytes")
	scan++
	err = (&scan).Scan(vlast)
	assert.NoError(t, err, "scan stringer")
	assert.Equal(t, vlast, scan, "scan stringer")
	scan++
	err = (&scan).Scan(int(vlast))
	assert.Error(t, err, "scan int")
	assert.NoError(t, (&scan).Scan(nil), "scan nil")
	assert.Error(t, err, (&scan).Scan("als;djfa;dlfja;ldfjal;sdjfl;sjdf;"), "invalid scan string")
}
