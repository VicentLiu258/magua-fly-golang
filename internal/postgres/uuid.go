package postgres

// nullUUID returns nil if s is nil or points to empty string, otherwise *s.
// Use for nullable UUID columns in INSERT/UPDATE.
func nullUUID(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
