package git_tfs_bridge

import (
	"time"
	"strings"
)

var russian2englishMonths = map[string]string {
	"января": "January",
	"февраля": "February",
	"марта": "March",
	"апреля": "April",
	"мая": "May",
	"июня": "June",
	"июля": "July",
	"августа": "August",
	"сентября": "September",
	"октября": "October",
	"ноября": "November",
	"декабря": "December",
}

func parseMaybeRussianDate(layout string, value string) (time.Time, error) {
	value = strings.Replace(value, " г.", "", 1)
	for rus, eng := range russian2englishMonths {
		if strings.Contains(value, rus) {
			value = strings.Replace(value, rus, eng, 1)
			break
		}
	}
	return time.Parse(layout, value)
}