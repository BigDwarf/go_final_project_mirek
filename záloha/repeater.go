package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const layout = "20060102"

func stringToTime(dateString string, layout string) (time.Time, error) {
	parsedDate, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}

func NextDate(now string, date string, repeat string) (string, error) {
	nowTimeTime, err := stringToTime(now, layout)
	if err != nil {
		return "", fmt.Errorf("invalid date: %s", now)
	}
	startDateTimeTime, err := stringToTime(date, layout)
	if err != nil {
		return "", fmt.Errorf("invalid date: %s", date)
	}
	codeAndNumber := strings.Split(repeat, " ")
	if len(codeAndNumber) == 0 || (codeAndNumber[0] != "y" && codeAndNumber[0] != "d") {
		return "", fmt.Errorf("invalid repeat code: %s", repeat)
	}

	if codeAndNumber[0] == "y" {

		nextTime := startDateTimeTime.AddDate(1, 0, 0)
		nextTimeString := nextTime.Format("20060102")
		return nextTimeString, nil
	}

	if codeAndNumber[0] == "d" {
		if len(codeAndNumber) != 2 {
			return "", fmt.Errorf("invalid day repeat format: %s", repeat)
		}

		i, err := strconv.Atoi(codeAndNumber[1])
		if err != nil || i > 400 {
			fmt.Println("Number higher than 400 or Error converting string to int:", err)
		}

		diff := nowTimeTime.Sub(startDateTimeTime)
		diffDays := int(diff.Hours() / 24)
		previousRepeat := startDateTimeTime.AddDate(0, 0, (diffDays/i)*i)
		nextTime := previousRepeat.AddDate(0, 0, i)
		nextTimeString := nextTime.Format("20060102")
		return nextTimeString, nil
	}

	return "", fmt.Errorf("unknown repeat code: %s", codeAndNumber[0])
}
