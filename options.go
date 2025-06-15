package csvamp

// IgnoreUnknownFieldNames is an option that can be passed to NewMapper / MustNewMapper
//
// if set to true, the mapper will ignore references to unknown CSV field (header) names
//
// By default, if a struct field references (via tag) an unknown CSV field (header) name - it will error when reading
type IgnoreUnknownFieldNames bool

// DefaultEmptyValues is an option that can be passed to NewMapper / MustNewMapper
//
// if set to true, when reading, empty fields in the CSV are treated as zero values for types bool, int, uint and float
//
// By default, reading empty CSV fields into bool, int, uint & float will cause an error
type DefaultEmptyValues bool
