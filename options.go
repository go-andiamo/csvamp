package csvamp

// IgnoreUnknownFieldNames is an option that can be passed to NewMapper/MustNewMapper
//
// # If set to true, the mapper will ignore references to unknown CSV field (header) names
//
// By default, if a struct field references (via tag) an unknown CSV field (header) name - it will error when reading
type IgnoreUnknownFieldNames bool
