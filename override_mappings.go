package csvamp

type OverrideMappings []OverrideMapping

type OverrideMapping struct {
	// FieldName is the name of the field in the original struct
	FieldName string
	// CsvFieldIndex is the field index in the CSV
	//
	// If this value is 0 (zero), the index is ignored and CsvFieldName is used instead
	//
	// If this value is negative, the index mapping is removed
	CsvFieldIndex int
	// CsvFieldName is the field name (header) in the CSV
	//
	// If this value is empty, the name is ignored and CsvFieldIndex is used instead
	//
	// If this value starts with a "-", the named mapping is removed
	CsvFieldName string
}
