/*
* @Author: wangqilong
* @Description:
* @File: isourl
* @Date: 2021/7/7 10:16 下午
 */

package server

import (
	"encoding/json"
	. "ipaas_bwstress/util/log"
	"ipaas_bwstress/util/request"
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

type UrlInfo struct {
	Url string `json:"url"`
}

type RespBody struct {
	Code int       `json:"code"`
	Data []UrlInfo `json:"data"`
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
	return
}

func UpdateResource() {
	go func() {
		for {
			GetUrls()
			time.Sleep(10 * time.Minute)
		}
	}()
}
