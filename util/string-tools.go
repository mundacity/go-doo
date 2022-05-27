package util

import (
	"fmt"
	"strings"
	"time"
)

func StringFromSlice(sl []string) string {
	bodyStr := ""
	for _, s := range sl {
		bodyStr += s + " "
	}
	return strings.Trim(bodyStr, " ")
}

func StringFromDate(d time.Time) string {
	m := int(d.Month())
	var mStr, dStr string
	if m < 10 {
		mStr = "0" + fmt.Sprint(m)
	} else {
		mStr = fmt.Sprint(m)
	}
	if d.Day() < 10 {
		dStr = "0" + fmt.Sprint(d.Day())
	} else {
		dStr = fmt.Sprint(d.Day())
	}
	final := fmt.Sprintf("%v-%v-%v", d.Year(), mStr, dStr)

	return final
}

func EscapeQuotes(input string) string {
	ret := strings.ReplaceAll(input, "\"", "``")
	ret = strings.ReplaceAll(ret, "'", "`")
	return ret
}

func UnEscapeQuotes(input string) string {
	ret := strings.ReplaceAll(input, "``", "\"")
	ret = strings.ReplaceAll(ret, "`", "'")
	return ret
}
