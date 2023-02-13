package xopat

func (t AttributeType) SpanAttributeType() AttributeType {
	switch t {
	case
		AttributeTypeInt,
		AttributeTypeInt8,
		AttributeTypeInt16,
		AttributeTypeInt32,
		AttributeTypeDuration:
		return AttributeTypeInt64
	case
		AttributeTypeFloat32:
		return AttributeTypeFloat64
	default:
		return t
	}
}
