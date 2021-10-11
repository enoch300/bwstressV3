/*
* @Author: wangqilong
* @Description:
* @File: log
* @Date: 2021/9/15 11:05 下午
 */

package log

import (
	"github.com/enoch300/glog"
	"github.com/sirupsen/logrus"
)

var L *logrus.Logger

func New(logPath string, app string, save uint) {
	L = glog.NewLogger(logPath, app, save)
}
