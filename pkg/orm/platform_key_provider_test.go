package orm

import (
	"testing"
)

func Test_generateCode(t *testing.T) {
	codes := make(map[string]int)
	for i := 0; i < 10000000; i++ {
		code := generateCode()
		codes[code]++
	}

	for _, code := range codes {
		if code > 1 {
			t.FailNow()
		}
	}
}