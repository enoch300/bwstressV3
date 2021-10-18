/*
* @Author: wangqilong
* @Description:
* @File: crontab
* @Date: 2021/9/16 11:24 上午
 */

package crontab

import (
	c "github.com/robfig/cron/v3"
	"ipaas_bwstress/util/config"
	. "ipaas_bwstress/util/log"
)

var crontab *c.Cron
var CurlStartEventCh = make(chan struct{})
var CurlStopEventCh = make(chan struct{})
var BTStartEventCh = make(chan struct{})
var BTStopEventCh = make(chan struct{})

func ServeStart() {
	L.Infof("Start working...")
	CurlStartEventCh <- struct{}{}
	BTStartEventCh <- struct{}{}
}

func ServeStop() {
	L.Infof("Stop working...")
	CurlStopEventCh <- struct{}{}
	BTStopEventCh <- struct{}{}
}

func UpdateCrontabJob() {
	for range config.UpdateEvent {
		CleanCrontabJob()
		if len(config.LocalCfg.Crontab) == 0 {
			L.Infof("no such crontab")
			return
		}

		for _, s := range config.LocalCfg.Crontab {
			AddCrontabJob(s.Start, ServeStart)
			L.Infof("Crontab add startJob: %v", s.Start)
			AddCrontabJob(s.Stop, ServeStop)
			L.Infof("Crontab add stopJob: %v", s.Stop)
		}
	}
}

func AddCrontabJob(sch string, fn func()) {
	_, err := crontab.AddFunc(sch, fn)
	if err != nil {
		L.Errorf("crontab.AddFunc: %v", err)
	}
}

func CleanCrontabJob() {
	var ids []c.EntryID
	for _, entry := range crontab.Entries() {
		crontab.Remove(entry.ID)
		ids = append(ids, entry.ID)
	}
	L.Infof("Crontab cleanJob: %v", ids)
}

func StartCrontab() {
	crontab = c.New()
	crontab.Start()
	go UpdateCrontabJob()
}
