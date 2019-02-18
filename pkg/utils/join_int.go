package utils

import (
	"strconv"
	"strings"
)

func JoinInt(list []int64, sep string) string {
	strs := []string{}
	for _, i := range list {
		strs = append(strs, strconv.FormatInt(i, 10))
	}
	return strings.Join(strs, sep)
}