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
	. "ipaas_bwstress/util/log"
	"strings"
)

var tFiles = []string{
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
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
}

func AppendTask(ip string) {
	isIPV6 := strings.Contains(ip, ":")
	if isIPV6 {
		return
	}

	for _, f := range tFiles {
		go func(tFile, srcIp string) {
			defer func() {
				if err := recover(); err != nil {
					L.Errorf("task panic to %v, %v", srcIp, err)
				}
			}()

			tf, err := torrentfile.Open("torrentfile/" + tFile)
			if err != nil {
				L.Errorf("torrentfile.Open: %v", err)
				return
			}

			if err = tf.DownloadToFile(srcIp); err != nil {
				L.Errorf("tf.DownloadToFile: %v", err.Error())
				return
			}
		}(f, ip)
	}

	L.Infof("Append task >>> srcIp: %v", ip)
}

func Working() {
	for _, outEthIfI := range collect.Net.IfiMap {
		AppendTask(outEthIfI.Ip)
	}
}
