package xopconst

import (
	"os"
	"path"
	"sync"
)

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
) *SpanType {
	if namespace == "" {
		namespace = path.Base(os.Args[0])
	}
	key := spanTypeKey{
		namespace: namespace,
		spanType:  spanType,
	}
	fmv := make(map[string]struct{})
	for _, f := range fieldsWithMultipleValues {
		fmv[f] = struct{}{}
	}
	st, ok := registeredSpanTypes.Load(key)
	if ok {
		return st.(*SpanType)
	}
	st, _ = registeredSpanTypes.LoadOrStore(key, &SpanType{
		key:                      key,
		fieldsToIndex:            fieldsToIndex,
		fmvMap:                   fmv,
		fieldsWithMultipleValues: fieldsWithMultipleValues,
		sent:                     0,
	})
	return st.(*SpanType)
}

func (st SpanType) Namespace() string                  { return st.key.namespace }
func (st SpanType) SpanType() string                   { return st.key.spanType }
func (st SpanType) FieldsToIndex() []string            { return st.fieldsToIndex }
func (st SpanType) FieldsWithMultipleValues() []string { return st.fieldsWithMultipleValues }
func (st SpanType) IsMultipleValued(f string) bool     { _, ok := st.fmvMap[f]; return ok }
