package lib

func DefaultString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
