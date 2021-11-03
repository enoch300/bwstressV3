/*
* @Author: wangqilong
* @Description:
* @File: curl
* @Date: 2021/10/12 6:36 下午
 */

package bwstress

import (
	"encoding/json"
	"fmt"
	gcurl "github.com/andelf/go-curl"
	"ipaas_bwstress/util"
	"ipaas_bwstress/util/collect"
	"ipaas_bwstress/util/config"
	"ipaas_bwstress/util/crontab"
	. "ipaas_bwstress/util/log"
	"ipaas_bwstress/util/request"
	"math"
	"math/rand"
	"strings"
	"time"
)

var ISOURL = []string{
	"https://mirrors.aliyun.com/centos/8.4.2105/isos/x86_64/CentOS-8.4.2105-x86_64-dvd1.iso",
	"https://mirrors.aliyun.com/centos/7.9.2009/isos/x86_64/CentOS-7-x86_64-Everything-2009.iso",
	"https://mirrors.aliyun.com/centos/8-stream/isos/x86_64/CentOS-Stream-8-x86_64-latest-dvd1.iso",
	"https://mirrors.aliyun.com/centos/8-stream/isos/x86_64/CentOS-Stream-8-x86_64-20210819-dvd1.iso",
	"https://mirrors.aliyun.com/centos/8.4.2105/isos/ppc64le/CentOS-8.4.2105-ppc64le-dvd1.iso",
	"https://mirrors.aliyun.com/centos/8.4.2105/isos/aarch64/CentOS-8.4.2105-aarch64-dvd1.iso",
	"http://mirrors.aliyun.com/centos/8.4.2105/isos/x86_64/CentOS-8.4.2105-x86_64-dvd1.iso",
	"http://mirrors.163.com/centos/8/isos/x86_64/CentOS-8.4.2105-x86_64-dvd1.iso",
	"http://mirrors.163.com/centos/8.4.2105/isos/x86_64/CentOS-8.4.2105-x86_64-dvd1.iso",
	"http://mirrors.163.com/centos/8.4.2105/isos/ppc64le/CentOS-8.4.2105-ppc64le-dvd1.iso",
	"http://mirrors.163.com/centos/8.4.2105/isos/aarch64/CentOS-8.4.2105-aarch64-dvd1.iso",
}

var ServeRun = false

type UrlInfo struct {
	Url string `json:"url"`
}

type RespBody struct {
	Code int       `json:"code"`
	Data []UrlInfo `json:"data"`
}

var userAgent = []string{
	"curl/7.29.0",
	"Wget/1.14 (linux-gnu)",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36",
}

type Task struct {
	Curl       *gcurl.CURL
	Url        string
	EthName    string
	Run        bool
	SpeedPoint []float64
	LimitRate  float64
}

func (t *Task) progress(float64, float64, float64, float64, interface{}) bool {
	return ServeRun && t.Run
}

func (t *Task) Do() {
	defer t.Curl.Cleanup()
	defer func() {
		if err := recover(); err != nil {
			L.Errorf("Do_panic: %v", err)
		}
	}()
	go t.SpeedMonitor()
	if err := t.Curl.Perform(); err != nil {
		if !strings.Contains(err.Error(), "Operation was aborted by an application callback") &&
			!strings.Contains(err.Error(), "Transferred a partial file") &&
			!strings.Contains(err.Error(), "Failure when receiving data from the peer") {
			L.Errorf("Perform ethName: %v, url: %v, %v", t.EthName, t.Url, err)
		}
	}
	t.Stop()
	return
}

func (t *Task) Stop() {
	t.Run = false
}

func (t *Task) SpeedMonitor() {
	for {
		if !t.Run {
			L.Infof("EthName: %v speed monitor stop, task run flag is %v", t.EthName, t.Run)
			break
		}

		if !ServeRun {
			L.Infof("EthName: %v speed monitor stop, server run flag is %v", t.EthName, ServeRun)
			break
		}

		time.Sleep(time.Minute)

		speed, _ := t.Curl.Getinfo(gcurl.INFO_SPEED_DOWNLOAD)
		code, _ := t.Curl.Getinfo(gcurl.INFO_HTTP_CODE)
		t.SpeedPoint = append(t.SpeedPoint, util.FormatFloat64(speed.(float64)*8/1000/1000))

		if len(t.SpeedPoint) > 120 {
			oldSpeed := t.SpeedPoint
			t.SpeedPoint = oldSpeed[1:len(t.SpeedPoint)]
			currentRate := util.FormatFloat64(util.Avg(t.SpeedPoint))
			if t.LimitRate == 30 && currentRate < 5 {
				t.Stop()
				L.Infof("EthName: %v, limitRate: %vMbps, currentRate: %vMbps, code: %v, rate slowly exit task", t.EthName, t.LimitRate, currentRate, code)
				break
			}

			if t.LimitRate == 10 && currentRate < 1 {
				t.Stop()
				L.Infof("EthName: %v, limitRate: %vMbps, currentRate: %vMbps, code: %v, rate slowly exit task", t.EthName, t.LimitRate, currentRate, code)
				break
			}
		}
	}
}

func GlobalInit() error {
	if err := gcurl.GlobalInit(gcurl.GLOBAL_ALL); err != nil {
		L.Errorf("gcurl.GlobalInit %v", err.Error())
		return err
	}
	L.Infof("gcurl GlobalInit")
	return nil
}

func GlobalCleanup() {
	gcurl.GlobalCleanup()
	L.Infof("gcurl GlobalCleanup")
}

func NewTask(ethName string, LimitRateMbps float64, url string) *Task {
	i := rand.Intn(len(userAgent))
	agent := userAgent[i]

	t := &Task{
		Url:       url,
		Curl:      gcurl.EasyInit(),
		Run:       true,
		EthName:   ethName,
		LimitRate: LimitRateMbps,
	}

	maxDownload := uint64(LimitRateMbps * 131072)
	t.Curl.Setopt(gcurl.OPT_URL, url)
	t.Curl.Setopt(gcurl.OPT_INTERFACE, t.EthName)
	t.Curl.Setopt(gcurl.OPT_USERAGENT, agent)
	t.Curl.Setopt(gcurl.OPT_NOPROGRESS, 0)
	t.Curl.Setopt(gcurl.OPT_FORBID_REUSE, 1)
	t.Curl.Setopt(gcurl.OPT_WRITEFUNCTION, func(buf []byte, userdata interface{}) bool {
		return true
	})
	t.Curl.Setopt(gcurl.OPT_CONNECTTIMEOUT, 5)
	//t.Curl.Setopt(gcurl.OPT_SSL_VERIFYPEER, false)
	t.Curl.Setopt(gcurl.OPT_PROGRESSFUNCTION, t.progress)
	t.Curl.Setopt(gcurl.OPT_MAX_RECV_SPEED_LARGE, maxDownload) //单位B/s
	return t
}

func GetUrls() {
	var respBody RespBody
	body, httpCode, err := request.Get("https://ipaas.paigod.work/api/v1/urlsource")
	if err != nil && httpCode != 200 {
		L.Errorf("fetch urls httpCode: %v, %v", httpCode, err)
		return
	}

	if err = json.Unmarshal(body, &respBody); err != nil {
		L.Errorf("fetch urls json.Unmarshal: %v, body: %v", err, string(body))
		return
	}

	if respBody.Code != 0 {
		L.Errorf("fetch urls respCode: %v, %v", respBody.Code, string(body))
		return
	}

	var urls []string
	if len(respBody.Data) > 0 {
		u := strings.Split(respBody.Data[0].Url, "\n")
		for _, url := range u {
			if url == "" {
				continue
			}
			urls = append(urls, strings.TrimSpace(url))
		}
	}

	ISOURL = urls
	L.Infof("fetch urls success count: %v", len(ISOURL))
	return
}

func UpdateResource() {
	for {
		GetUrls()
		time.Sleep(10 * time.Minute)
	}
}

func GetHealthResource(ethName string) (url string, err error) {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(userAgent))
	agent := userAgent[i]

	var healthList []string
	for _, u := range ISOURL {
		t := gcurl.EasyInit()
		t.Setopt(gcurl.OPT_URL, u)
		t.Setopt(gcurl.OPT_RANGE, "0-1024")
		t.Setopt(gcurl.OPT_USERAGENT, agent)
		t.Setopt(gcurl.OPT_INTERFACE, ethName)
		t.Setopt(gcurl.OPT_NOPROGRESS, 0)
		t.Setopt(gcurl.OPT_FORBID_REUSE, 1)
		t.Setopt(gcurl.OPT_WRITEFUNCTION, func(buf []byte, userdata interface{}) bool {
			return true
		})
		t.Setopt(gcurl.OPT_CONNECTTIMEOUT, 5)
		//t.Curl.Setopt(gcurl.OPT_SSL_VERIFYPEER, false)

		t.Setopt(gcurl.OPT_PROGRESSFUNCTION, func(float64, float64, float64, float64, interface{}) bool {
			return true
		})

		if err = t.Perform(); err != nil {
			t.Cleanup()
			continue
		}
		code, _ := t.Getinfo(gcurl.INFO_HTTP_CODE)
		if code.(int) == 206 {
			healthList = append(healthList, u)
		}
		t.Cleanup()
	}

	if len(healthList) == 0 {
		return "", fmt.Errorf("has no health resource")
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(healthList))
	u := healthList[n]
	return u, nil
}

func AppendCurlTask(ethName string, n int, LimitRateMbps float64, outEthIfRecv float64, outIfMaxRecvBw float64) {
	for i := 0; i < n; i++ {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(ISOURL))
		t := NewTask(ethName, LimitRateMbps, ISOURL[index])
		go t.Do()
	}
	L.Infof("Add curl task >>> ethName: %v %v, recv: %vMbps,  maxRecv: %vMbps", ethName, n, outEthIfRecv, outIfMaxRecvBw)
}

func CURLWorking() {
	L.Infof("+++ Third-party start +++")
	if err := GlobalInit(); err != nil {
		L.Errorf("Curl GlobalInit %v", err.Error())
		return
	}
	defer func() {
		GlobalCleanup()
	}()
	go UpdateResource()

	for {
		time.Sleep(5 * time.Minute)
		if !ServeRun {
			time.Sleep(2 * time.Minute)
			return
		}

		maxDownload := float64(config.LocalCfg.MaxDownload)
		maxRecvSendRate := float64(config.LocalCfg.RecvSendRate)

		//maxRecvSendBwPer 小于等于0 表示不判断总下行占总上行比例, 否则需要判断.
		if (maxRecvSendRate <= 0 && collect.OutEthRecvByteAvg < maxDownload) || (maxRecvSendRate > 0 && collect.OutEthRecvSendUseRate < maxRecvSendRate && collect.OutEthRecvByteAvg < maxDownload) {
			for ethName, outEthIfI := range collect.Net.IfiMap {
				outEthIfRecv := util.FormatFloat64(util.ByteToBitM(outEthIfI.RecvByteAvg))
				if outEthIfRecv < collect.OutIfMaxRecvBw {
					if collect.OutIfMaxRecvBw > 30 {
						n := int(math.Floor((collect.OutIfMaxRecvBw - outEthIfRecv) / 30))
						if n > 0 {
							AppendCurlTask(ethName, 3, 30, outEthIfRecv, collect.OutIfMaxRecvBw)
						}
						//AppendCurlTask(ethName, 5, 30, outEthIfRecv, collect.OutIfMaxRecvBw)
					} else {
						n := int(math.Floor((collect.OutIfMaxRecvBw - outEthIfRecv) / 10))
						if n > 0 {
							AppendCurlTask(ethName, 3, 10, outEthIfRecv, collect.OutIfMaxRecvBw)
						}
						//AppendCurlTask(ethName, 5, 10, outEthIfRecv, collect.OutIfMaxRecvBw)
					}
				}
			}
		}
	}
}

func CurlRun() {
	for {
		select {
		case <-crontab.CurlStartEventCh:
			ServeRun = true
			go CURLWorking()
		case <-crontab.CurlStopEventCh:
			L.Infof("+++ Third-party stop +++")
			ServeRun = false
		}
	}
}
