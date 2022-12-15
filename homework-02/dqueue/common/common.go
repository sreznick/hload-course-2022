package common

func Fmap(arr []string, f func(string) (int, error)) []int {
	var ans []int
	for _, s := range arr {
		i, err := f(s)
		if err != nil {
			continue
		}

		ans = append(ans, i)
	}

	return ans
}

func Min(arr []int) (int, int) {
	min := arr[0]
	id := 0
	for i, e := range arr[1:] {
		if min > e {
			min = e
			id = i
		}
	}

	return min, id
}
