package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	hourDayNum   = 24
	hourMonthNum = hourDayNum * 30
	hourYearNum  = hourMonthNum * 12
)

var (
	subsVarDetectorRe    = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	renewBeforeRe        = regexp.MustCompile(`^(\d+)(d|M|Y)$`)
	customDurationMapper = map[string]int{
		"d": hourDayNum,
		"M": hourMonthNum,
		"Y": hourYearNum,
	}
)

func durationParse(input string) (time.Duration, error) {
	match := renewBeforeRe.FindStringSubmatch(input)
	if len(match) != 3 {
		return 0, fmt.Errorf("%s is not match with duration pattern", input)
	}

	// ignore the error, since in previous regex match we sure it's a number
	numMatch, _ := strconv.Atoi(match[1])
	convertedNumHour := customDurationMapper[match[2]] * numMatch
	finalHourDurationStr := fmt.Sprintf("%dh", convertedNumHour)
	res, err := time.ParseDuration(finalHourDurationStr)
	if err != nil {
		return 0, fmt.Errorf("error during parse duration string: %v", err)
	}

	return res, nil
}

func evalEnvVar(input string) string {
	matches := subsVarDetectorRe.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return input
	}

	for _, match := range matches {
		envVar := match[1]

		envValue, _ := os.LookupEnv(envVar)
		input = strings.ReplaceAll(input, fmt.Sprintf("${%s}", envVar), envValue)
	}
	return input
}

// Contains checks if a value exists in a slice
func contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
