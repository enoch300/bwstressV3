/*
* @Author: wangqilong
* @Description:
* @File: crontab
* @Date: 2021/9/16 11:24 上午
 */

package crontab

import (
	c "github.com/robfig/cron/v3"
	"ipaas_bwstress/bt/torrentfile"
	"ipaas_bwstress/server/bwstress"
	"ipaas_bwstress/util/config"
	. "ipaas_bwstress/util/log"
)

var crontab *c.Cron

func ServeStart() {
	bwstress.Working()
	L.Infof("Start working...")
}

func ServeStop() {
	close(torrentfile.DoneCh)
	L.Infof("Stop working...")
}

func UpdateCrontabJob() {
	for range config.UpdateEvent {
		CleanCrontabJob()
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
