package xopconst

import (
	"os"
	"sync"
	"path"
)

type SpanType struct {
	key                      spanTypeKey
	fieldsToIndex            []string
	fmvMap                   map[string]struct{}
	fieldsWithMultipleValues []string
	sent                     int32 // 1 = sent, 0 = not sent
}

var SubspanType = RegisterSpanType(path.Base(os.Args[0]), "subspan", nil, nil)

var spanTypeKey struct {
	namespace string
	spanType  string
}

var registeredSpanTypes sync.Map

func RegisterSpanType(
	namespace string, // program name
	spanType string,
	fieldsToIndex []string,
	fieldsWithMultipleValues []string) SpanType {
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
	st, _ = registeredSpanTypes.LoadOrStore(key, &SpanType{
		key:                      key,
		fieldsToIndex:            fieldsToIndex,
		fmvMap:                   fmv,
		fieldsWithMultipleValues: fieldsWithMultipleValues,
		sent:                     0,
	})
	return st
}

func (st SpanType) Namespace() string                  { return st.key.namespace }
func (st SpanType) SpanType() string                   { return st.key.spanType }
func (st SpanType) FieldsToIndex() []string            { return st.fieldsToIndex }
func (st SpanType) FieldsWithMultipleValues() []string { return st.fieldsWithMultipleValues }
func (st SpanType) IsMultipleValued(f name) bool       { _, ok := st.fmv[f]; return ok }
