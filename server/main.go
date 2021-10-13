/*
* @Author: wangqilong
* @Description:
* @File: main
* @Date: 2021/9/15 11:18 下午
 */

package server

import (
	"ipaas_bwstress/server/bwstress"
	"ipaas_bwstress/util"
	"ipaas_bwstress/util/collect"
	"ipaas_bwstress/util/config"
	"ipaas_bwstress/util/crontab"
	. "ipaas_bwstress/util/log"
)

func Run() {
	L.Info("Server start")
	util.GetMachineId()
	config.CheckRemoteConf()
	crontab.StartCrontab()
	collect.NetCollect()
	go bwstress.CurlRun()
	go bwstress.BTRun()
	select {}
}
