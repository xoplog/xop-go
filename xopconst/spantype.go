package xopconst

import (
	"os"
	"path"
	"sync"
)

var AllSpans = RegisterSpanType("xop", "request",
	[]string{"xop:type", "xop:is-request"},
	nil)

type SpanType struct {
	key                      spanTypeKey
	fieldsToIndex            []string
	fmvMap                   map[string]struct{}
	fieldsWithMultipleValues []string
	sent                     int32 // 1 = sent, 0 = not sent
}

var SubspanType = RegisterSpanType(path.Base(os.Args[0]), "subspan", nil, nil)

type spanTypeKey struct {
	namespace string
	spanType  string
}

var registeredSpanTypes sync.Map

func RegisterSpanType(
	namespace string, // program name
	spanType string,
	fieldsToIndex []string,
	fieldsWithMultipleValues []string,
	baseSpanTypes ...SpanType,
) SpanType {
	if namespace == "" {
		namespace = path.Base(os.Args[0])
	}
	key := spanTypeKey{
		namespace: namespace,
		spanType:  spanType,
	}
	fieldsToIndex = combineValues(fieldsToIndex, baseSpanTypes, func(st SpanType) []string {
		return st.fieldsToIndex
	})
	fieldsWithMultipleValues = combineValues(fieldsWithMultipleValues, baseSpanTypes, func(st SpanType) []string {
		return st.fieldsWithMultipleValues
	})

	fmv := make(map[string]struct{})
	for _, f := range fieldsWithMultipleValues {
		fmv[f] = struct{}{}
	}
	st, ok := registeredSpanTypes.Load(key)
	if ok {
		return *(st.(*SpanType))
	}
	st, _ = registeredSpanTypes.LoadOrStore(key, &SpanType{
		key:                      key,
		fieldsToIndex:            fieldsToIndex,
		fmvMap:                   fmv,
		fieldsWithMultipleValues: fieldsWithMultipleValues,
		sent:                     0,
	})
	return *(st.(*SpanType))
}

func (st SpanType) Namespace() string                  { return st.key.namespace }
func (st SpanType) SpanType() string                   { return st.key.spanType }
func (st SpanType) FieldsToIndex() []string            { return st.fieldsToIndex }
func (st SpanType) FieldsWithMultipleValues() []string { return st.fieldsWithMultipleValues }
func (st SpanType) IsMultipleValued(f string) bool     { _, ok := st.fmvMap[f]; return ok }

func combineValues(startingSet []string, otherTypes []SpanType, accessor func(SpanType) []string) []string {
	n := make([]string, 0, len(startingSet)*(1+len(otherTypes)))
	seen := make(map[string]struct{})
	for _, v := range startingSet {
		if _, ok := seen[v]; !ok {
			n = append(n, v)
			seen[v] = struct{}{}
		}
	}
	for _, spanType := range otherTypes {
		for _, v := range accessor(spanType) {
			if _, ok := seen[v]; !ok {
				n = append(n, v)
				seen[v] = struct{}{}
			}
		}
	}
	return n
}
