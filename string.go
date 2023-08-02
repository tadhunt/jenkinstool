package jenkinstool

func String(s *string) string {
	if s == nil {
		return "<nil>"
	}

	return *s
}
