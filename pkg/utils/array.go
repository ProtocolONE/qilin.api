package utils

func Contains(arr []string, item string) bool {
	for _, cur := range arr {
		if item == cur {
			return true
		}
	}

	return false
}
