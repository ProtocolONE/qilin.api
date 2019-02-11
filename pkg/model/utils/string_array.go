package utils

import "strings"

type StringArray []string

//Contains is func for searching string in array of strings
func (arr StringArray) Contains(search string) bool {
	for _, current := range arr {
		if current == search {
			return true
		}
	}
	return false
}

func (arr StringArray) String() string {
	return strings.Join(arr, ", ")
}

