package main

import (
	"strconv"
	"strings"
)

func RemoveEmpty(strings []string) []string {
	noEmpty := []string{}
	for i := range strings {
		s := strings[i]
		if len(s) != 0 {
			noEmpty = append(noEmpty, s)
		}
	}
	return noEmpty
}

func TimeToInt(time string) int {
	hour := strings.Split(time, ":")[0]
	i, err := strconv.Atoi(hour)
	if err != nil {
		return -1
	}
	return i
}
