/*
* @Author: wangqilong
* @Description:
* @File: config
* @Date: 2021/9/15 11:35 下午
 */

package config

import (
	"encoding/json"
	"fmt"
	"github.com/golang-module/carbon"
	"ipaas_bwstress/util"
	. "ipaas_bwstress/util/log"
	"ipaas_bwstress/util/request"
	"time"
)

var API string
var LocalCfg *LocalConf
var UpdateEvent = make(chan struct{})

type Crontab struct {
	Start string `yaml:"start"`
	Stop  string `yaml:"stop"`
}

type RemoteConf struct {
	MachineId    string    `json:"machine_id"`
	RecvSendRate int64     `json:"recv_send_rate"`
	MaxDownload  int64     `json:"max_download"`
	MaxUpload    int64     `json:"max_upload"`
	Enabled      int       `json:"enabled"`
	Filter       string    `json:"filter"`
	Crontab      []Crontab `json:"crontab"`
	UpdateAt     string    `json:"updated_at"`
}

type LocalConf struct {
	MachineId    string    `json:"machine_id"`
	RecvSendRate int64     `json:"recv_send_rate"`
	MaxDownload  int64     `json:"max_download"`
	MaxUpload    int64     `json:"max_upload"`
	Enabled      int       `json:"enabled"`
	Filter       string    `json:"filter"`
	Crontab      []Crontab `json:"crontab"`
	UpdateAt     string    `json:"updated_at"`
}

func (l *LocalConf) IsModify(r *RemoteConf) (bool, error) {
	isModify := carbon.Parse(r.UpdateAt).Gt(carbon.Parse(l.UpdateAt))
	return isModify, nil
}

func (l *LocalConf) Update(r *RemoteConf) {
	before, _ := json.Marshal(l)
	L.Infof("Before update config: %v", string(before))
	l.Filter = r.Filter
	l.Crontab = r.Crontab
	l.Enabled = r.Enabled
	l.UpdateAt = r.UpdateAt
	l.MachineId = r.MachineId
	l.MaxUpload = r.MaxUpload
	l.MaxDownload = r.MaxDownload
	l.RecvSendRate = r.RecvSendRate
	//crontab.UpdateCrontabJob(l.Crontab)
	UpdateEvent <- struct{}{}
	after, _ := json.Marshal(l)
	L.Infof("After update config: %v", string(after))
}

type Response struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data []*RemoteConf `json:"data"`
}

func FetchRemoteConf() (remoteConf []*RemoteConf, err error) {
	body, code, err := request.Get(API + "?machine_id=" + util.MachineID)
	if err != nil {
		return remoteConf, err
	}

	if code != 200 {
		return remoteConf, fmt.Errorf("fetch httpCode: %v, body: %v", code, string(body))
	}

	var response Response
	if err = json.Unmarshal(body, &response); err != nil {
		return remoteConf, err
	}

	if response.Code == 1 {
		return remoteConf, fmt.Errorf(response.Msg)
	}

	return response.Data, nil
}

func (r *RemoteConf) Check() {
	go func() {
		for {
			time.Sleep(time.Minute)
			remoteConfData, err := FetchRemoteConf()
			if err != nil {
				L.Errorf("FetchRemoteConf: %v", err.Error())
				continue
			}

			if len(remoteConfData) == 0 {
				L.Warnf("%v has been deleted.", util.MachineID)
				if LocalCfg.Enabled == 1 {
					defaultConf := &RemoteConf{
						MachineId:    util.MachineID,
						RecvSendRate: 0,
						MaxDownload:  0,
						MaxUpload:    0,
						Enabled:      0,
						Filter:       "0.|127.|192.|172.",
						Crontab:      make([]Crontab, 0),
						UpdateAt:     "",
					}
					LocalCfg.Update(defaultConf)
				}
				continue
			}

			isModify, err := LocalCfg.IsModify(remoteConfData[0])
			if err != nil {
				L.Errorf("localConf.IsModify: %v", err.Error())
				continue
			}

			if isModify {
				LocalCfg.Update(remoteConfData[0])
			}
		}
	}()
}

func NewRemoteConf() *RemoteConf {
	return &RemoteConf{}
}

func CheckRemoteConf() {
	LocalCfg = &LocalConf{
		MachineId: util.MachineID,
		Enabled:   0,
	}

	remoteConf := NewRemoteConf()
	remoteConf.Check()
}
