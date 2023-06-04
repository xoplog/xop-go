package replayutil

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/xoplog/xop-go/xopat"
	"github.com/xoplog/xop-go/xopproto"
)

type GlobalAttributeDefinitions struct {
	CommonDefinitions map[string]*DecodeAttributeDefinition
	Requests          map[string]*RequestAttributeDefinitons
}

type RequestAttributeDefinitons struct {
	AttributeDefinitions       map[string]*DecodeAttributeDefinition
	GlobalAttributeDefinitions GlobalAttributeDefinitions
}

func NewGlobalAttributeDefinitions() *GlobalAttributeDefinitions {
	return &GlobalAttributeDefinitions{
		CommonDefinitions: make(map[string]*DecodeAttributeDefinition),
		Requests:          make(map[string]*RequestAttributeDefinitons),
	}
}

// NewRequestAttributeDefinitions is not thread-safe
func (g GlobalAttributeDefinitions) NewRequestAttributeDefinitions(requestID string) *RequestAttributeDefinitons {
	r := &RequestAttributeDefinitons{
		AttributeDefinitions:       make(map[string]*DecodeAttributeDefinition),
		GlobalAttributeDefinitions: g,
	}
	g.Requests[requestID] = r
	return r
}

type DecodeAttributeDefinition struct {
	xopat.Make
	AttributeType xopproto.AttributeType `json:"vtype"`
	SpanID        string                 `json:"span.id"`
}

// Lookup is not thread safe
func (r *RequestAttributeDefinitons) Lookup(key string) *DecodeAttributeDefinition {
	if d, ok := r.AttributeDefinitions[key]; ok {
		return d
	}
	return r.GlobalAttributeDefinitions.CommonDefinitions[key]
}

func (r *RequestAttributeDefinitons) Decode(inputText string) error {
	return r.GlobalAttributeDefinitions.Decode(inputText)
}

func (g *GlobalAttributeDefinitions) Lookup(requestID string, key string) *DecodeAttributeDefinition {
	r, ok := g.Requests[requestID]
	if !ok {
		return nil
	}
	return r.Lookup(key)
}

// Decode is not thread-safe
func (g *GlobalAttributeDefinitions) Decode(inputText string) error {
	var aDef DecodeAttributeDefinition
	err := json.Unmarshal([]byte(inputText), &aDef)
	if err != nil {
		return errors.Wrapf(err, "decode attribute defintion (%s)", inputText)
	}
	if aDef.SpanID != "" {
		r, ok := g.Requests[aDef.SpanID]
		if !ok {
			return errors.Errorf("attribute definition for %s references non-existant request %s", aDef.Key, aDef.SpanID)
		}
		r.AttributeDefinitions[aDef.Key] = &aDef
	} else {
		g.CommonDefinitions[aDef.Key] = &aDef
	}
	return nil
}
