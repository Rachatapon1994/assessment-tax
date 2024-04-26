package tax

import "math"

type Level struct {
	Name         string
	StartAmount  float64
	EndAmount    float64
	Percentage   float64
	MaxDeduction float64
}

func getLevels() []Level {
	return []Level{
		{"0-150,000", 0, 150000, 0, 0},
		{"150,001-500,000", 150001, 500000, 10, 35000},
		{"500,001-1,000,000", 500001, 1000000, 15, 75000},
		{"1,000,001-2,000,000", 1000001, 2000000, 20, 200000},
		{"2,000,001 ขึ้นไป", 2000001, math.MaxFloat64, 35, math.MaxFloat64},
	}
}
