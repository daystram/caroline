package util

func Plural(w string, n int) string {
	if n == 1 {
		return w
	}

	return w + "s"
}
