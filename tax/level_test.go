package tax

import (
	"math"
	"reflect"
	"testing"
)

func Test_getLevels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want []Level
	}{
		{"Should get taxes all levels correctly", []Level{
			{"0-150,000", 0, 150000, 0, 0},
			{"150,001-500,000", 150001, 500000, 10, 35000},
			{"500,001-1,000,000", 500001, 1000000, 15, 75000},
			{"1,000,001-2,000,000", 1000001, 2000000, 20, 200000},
			{"2,000,001 ขึ้นไป", 2000001, math.MaxFloat64, 35, math.MaxFloat64},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLevels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLevels() = %v, want %v", got, tt.want)
			}
		})
	}
}
