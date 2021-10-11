/*
* @Author: wangqilong
* @Description:
* @File: math
* @Date: 2021/9/16 10:39 上午
 */

package util

//平均
func Avg(l []float64) float64 {
	var sum float64
	for _, i := range l {
		sum += i
	}
	return sum / float64(len(l))
}
