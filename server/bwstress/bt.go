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

var tFiles = []string{
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"debian-11.1.0-i386-DVD-1.iso.torrent",
	"debian-11.1.0-i386-DVD-1.iso.torrent",
	"debian-11.1.0-i386-DVD-1.iso.torrent",
	"debian-11.1.0-i386-DVD-1.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
	"ubuntu-20.04.3-live-server-amd64.iso.torrent",
	"ubuntu-20.04.3-live-server-amd64.iso.torrent",
	"ubuntu-20.04.3-live-server-amd64.iso.torrent",
	"ubuntu-20.04.3-live-server-amd64.iso.torrent",
	"ubuntu-21.04-live-server-amd64.iso.torrent",
	"ubuntu-21.04-live-server-amd64.iso.torrent",
	"ubuntu-21.04-live-server-amd64.iso.torrent",
	"ubuntu-21.04-live-server-amd64.iso.torrent",
	"ubuntu-21.04-desktop-amd64.iso.torrent",
	"ubuntu-21.04-desktop-amd64.iso.torrent",
	"ubuntu-21.04-desktop-amd64.iso.torrent",
	"ubuntu-21.04-desktop-amd64.iso.torrent",
}

func AppendBtTask(ethName string) {
	ip := collect.Net.IfiMap[ethName].Ip
	isIPV6 := strings.Contains(ip, ":")
	if isIPV6 {
		return
	}

	for _, f := range tFiles {
		go func(tFile, ethName string) {
			defer func() {
				if err := recover(); err != nil {
					L.Errorf("Bt task panic to ethName %v ip %v, %v", ethName, ip, err)
				}
			}()

			tf, err := torrentfile.Open("../.torrentfile/" + tFile)
			if err != nil {
				L.Errorf("torrentfile.Open: %v", err)
				return
			}

			if err = tf.DownloadToFile(ethName, tFile); err != nil {
				L.Errorf("tf.DownloadToFile: %v", err.Error())
				return
			}
		}(f, ethName)
	}

	L.Infof("PreAdd %v tasks >>> ethName: %v, ip: %v", len(tFiles), ethName, ip)
}

func BTWorking() {
	torrentfile.DoneCh = make(chan struct{})
	for ethName, _ := range collect.Net.IfiMap {
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
			L.Infof("+++ p2p start +++")
			go BTWorking()
		case <-crontab.BTStopEventCh:
			L.Infof("+++ p2p stop +++")
			BTStop()
		}
	}
}
