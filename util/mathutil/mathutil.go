package mathutil

import "math"

// Sum return the summation value of float64 values.
func Sum(vals []float64) float64 {
	var sum float64
	for i := 0; i < len(vals); i++ {
		sum += vals[i]
	}
	return sum
}

// Average returns the mean value of float64 values.
// Returns zero if the vals length is 0.
func Average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	return Sum(vals) / float64(len(vals))
}

// StdDev returns the standard deviation of float64 values, with an input
// average.
// Returns zero if the vals length is 0.
func StdDev(vals []float64, avg float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < len(vals); i++ {
		dis := vals[i] - avg
		sum += dis * dis
	}
	return math.Sqrt(sum / float64(len(vals)))
}

// StdAverage return the pooled variance
func StdAverage(stds []float64, nums []int) float64 {
	var stdAll float64
	var num int
	for i := 0; i < len(stds); i++ {
		stdAll += float64(nums[i]-1) * stds[i] * stds[i]
		num += nums[i] - 1
	}
	if num == 0 {
		return 0
	}
	return math.Sqrt(stdAll / float64(num))
}

// Score returns the score of last value via 3-sigma,with an input avg and std.
//	states that nearly all values (99.7%) lie within the 3 standard deviations
//	of the mean in a normal distribution.
func Score(last float64, avg float64, std float64) float64 {
	var score float64
	if std == 0 { // Eadger
		switch {
		case last == avg:
			score = 0
		case last > avg:
			score = 1
		case last < avg:
			score = -1
		}
		return score
	}
	return (last - avg) / (3 * std) // 3-sigma
}
