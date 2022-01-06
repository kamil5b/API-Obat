package utils

import "strconv"

func MapStringToInt(data map[string]string) map[string]int {
	dataint := make(map[string]int)
	for x, y := range data {
		tmp, err := strconv.Atoi(y)
		if err == nil {
			dataint[x] = tmp
		}
	}
	return dataint
}
