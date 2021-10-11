/*
* @Author: wangqilong
* @Description:
* @File: util
* @Date: 2021/9/15 11:28 下午
 */

package util

import (
	"fmt"
	"strconv"
)

func FormatFloat64(f float64) float64 {
	v, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", f), 64)
	return v
}

func BitToByteM(f float64) float64 {
	return f / 8 / 1000 / 1000
}

func ByteToBitM(f float64) float64 {
	return f * 8 / 1000 / 1000
}

func StringsJoin(str []string, spe string) string {
	var r string
	for _, s := range str {
		r += fmt.Sprintf("%v%v", s, spe)
	}

	return r
}
