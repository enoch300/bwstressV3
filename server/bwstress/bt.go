/*
* @Author: wangqilong
* @Description:
* @File: bwstress
* @Date: 2021/9/16 2:14 上午
 */

package bwstress

import (
	"ipaas_bwstress/bt/torrentfile"
	"ipaas_bwstress/util/collect"
	"ipaas_bwstress/util/crontab"
	. "ipaas_bwstress/util/log"
	"strings"
)

func AppendBtTask(ethName string) {
	ip := collect.Net.IfiMap[ethName].Ip
	isIPV6 := strings.Contains(ip, ":")
	if isIPV6 {
		return
	}

	for _, f := range torrentfile.TFiles {
		for i := 0; i < 4; i++ {
			tf, err := torrentfile.Open("../.torrentfile/" + f)
			if err != nil {
				L.Errorf("EthName: %v, ip: %v, torrentfile.Open: %v", ethName, ip, err)
				return
			}

			tf.DownloadToFile(ethName, f)
			L.Infof("PreAdd tasks success >>> ethName: %v, ip: %v, tfile: %v", ethName, ip, f)
		}
	}
}

func BTWorking() {
	torrentfile.DoneCh = make(chan struct{})
	go torrentfile.RequestTFilesPeers()
	for ethName := range collect.Net.IfiMap {
		AppendBtTask(ethName)
	}
}

func BTStop() {
	close(torrentfile.DoneCh)
}

func BTRun() {
	for {
		select {
		case <-crontab.BTStartEventCh:
			L.Infof("+++ detect start +++")
			go BTWorking()
		case <-crontab.BTStopEventCh:
			L.Infof("+++ detect stop +++")
			BTStop()
		}
	}
}
