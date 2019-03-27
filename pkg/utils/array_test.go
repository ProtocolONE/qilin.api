package utils

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		arr  []string
		item string
		want bool
	}{
		{name: "Contains string", arr: []string{"1", "2", "test", "3"}, item: "test", want: true},
		{name: "Not contains string", arr: []string{"1", "2", "test", "3"}, item: "test2", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.arr, tt.item); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
