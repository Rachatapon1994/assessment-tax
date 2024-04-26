package tax

import "math"

type Level struct {
	startAmount  float64
	endAmount    float64
	percentage   float64
	maxDeduction float64
}

func getLevels() []Level {
	return []Level{
		{0, 150000, 0, 0},
		{150001, 500000, 10, 35000},
		{500001, 1000000, 15, 75000},
		{1000001, 2000000, 20, 200000},
		{2000001, math.MaxFloat64, 35, math.MaxFloat64},
	}
}
