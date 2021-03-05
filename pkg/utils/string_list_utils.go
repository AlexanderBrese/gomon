package utils

func Contains(list []string, element string) bool {
	for _, e := range list {
		if e == element {
			return true
		}
	}

	return false
}

func Size(list []string) int {
	size := 0

	for _, e := range list {
		if e != "" {
			size++
		}
	}

	return size
}
