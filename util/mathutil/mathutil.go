package mathutil

import "math"

// Sum return the summation value of float64 values.
func Sum(vals []float64) float64 {
	var total float64
	for i := 0; i < len(vals); i++ {
		total += vals[i]
	}
	return total
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
	var total float64
	for i := 0; i < len(vals); i++ {
		dis := vals[i] - avg
		total += dis * dis
	}
	return math.Sqrt(total / float64(len(vals)))
}

// StdAverage return the pooled variance
func StdAverage(stds []float64, nums []int) float64 {
	var stdTotal float64
	var num int
	for i := 0; i < len(stds); i++ {
		stdTotal += float64(nums[i]-1) * stds[i] * stds[i]
		num += nums[i] - 1
	}
	if num == 0 {
		return 0
	}
	return math.Sqrt(stdTotal / float64(num))
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

// Min returns the min value of float64 array
func Min(vals []float64) float64 {
	if len(vals) == 0 {
		return math.Inf(-1)
	}
	min := math.Inf(1)
	for _, val := range vals {
		if val < min {
			min = val
		}
	}
	return min
}

// Max returns the max value of float64 array
func Max(vals []float64) float64 {
	if len(vals) == 0 {
		return math.Inf(1)
	}
	max := math.Inf(-1)
	for _, val := range vals {
		if val > max {
			max = val
		}
	}
	return max
}

// Saturation returns val if min <= val <= max or
//    return max if val > max or
//    return min if val < min
func Saturation(val, from, to float64) float64 {
	max := math.Max(from, to)
	min := math.Min(from, to)
	if val > max {
		return max
	}
	if val < min {
		return min
	}
	return val
}

// AbsMin returns value with the min absolute value of float64 array
func AbsMin(vals []float64) float64 {
	if len(vals) == 0 {
		return math.Inf(-1)
	}
	min := math.Inf(1)
	for _, val := range vals {
		if math.Abs(val) < math.Abs(min) {
			min = val
		}
	}
	return min
}
