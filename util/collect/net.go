/*
* @Author: wangqilong
* @Description:
* @File: netcollect
* @Date: 2021/9/15 10:49 下午
 */

package collect

import (
	"encoding/json"
	"github.com/enoch300/collectd/net"
	. "ipaas_bwstress/util"
	"ipaas_bwstress/util/config"
	. "ipaas_bwstress/util/log"
	"ipaas_bwstress/util/request"
	"strings"
	"time"
)

var Net *net.NetWork
var OutEthRecvByteAvg float64
var OutEthSendByteAvg float64
var OutEthRecvSendUseRate float64
var OutIfMaxRecvBw float64
var OutIfMaxSendBw float64

const (
	BANDWIDTH_REPORT_API = "https://ipaas.paigod.work/api/v1/bandwidth"
)

type Bandwidth struct {
	EthName     string  `json:"eth_name"`
	MachineId   string  `json:"machine_id"`
	Upload      float64 `json:"upload"`
	Download    float64 `json:"download"`
	MaxUpload   float64 `json:"max_upload"`
	MaxDownload float64 `json:"max_download"`
}

func (b *Bandwidth) Report() {
	data, _ := json.Marshal(b)
	respBody, httpCode, err := request.Post(BANDWIDTH_REPORT_API, data)
	if err != nil {
		L.Errorf("Bandwidth report err: %v", err.Error())
		return
	}

	if httpCode != 200 {
		L.Errorf("Bandwidth report httpCode: %v", httpCode)
		return
	}

	var resp config.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		L.Errorf("Bandwidth report json.Unmarshal: %v", err.Error())
		return
	}

	if resp.Code != 0 {
		L.Errorf("Bandwidth report response code: %v, msg: %v", resp.Code, resp.Msg)
		return
	}

	L.Infof("Bandwidth report success")
}

func NetCollect() {
	Net = net.NewNetwork([]string{}, []string{"docker", "lo"}, []string{}, []string{})
	go func() {
		for {
			//更新忽略IP
			filterIps := strings.Split(config.LocalCfg.Filter, "|")
			Net.IgnoreIP = filterIps

			err := Net.Collect()
			if err != nil {
				L.Errorf("Net.Collect: %v", err.Error())
				time.Sleep(60 * time.Second)
				continue
			}

			OutEthRecvByteAvg = FormatFloat64(ByteToBitM(Net.OutEthRecvByteAvgFunc()))
			OutEthSendByteAvg = FormatFloat64(ByteToBitM(Net.OutEthSendByteAvgFunc()))
			OutEthRecvSendUseRate = FormatFloat64(OutEthRecvByteAvg / OutEthSendByteAvg * 100)
			OutIfMaxRecvBw = FormatFloat64(float64(config.LocalCfg.MaxDownload) / float64(len(Net.IfiNames))) //根据网卡数量平均下行带宽
			OutIfMaxSendBw = FormatFloat64(float64(config.LocalCfg.MaxUpload) / float64(len(Net.IfiNames)))   //根据网卡数量平均上行带宽
			L.Infof("Monitor interface: %v", Net.IfiNames)
			L.Infof("Monitor details >>> outMaxRecvBw: %vMbps, outMaxSendBw: %vMbps, outRecvTotal: %vMbps, outSendTotal: %vMbps, outRecv/outSend useRate: %v%%",
				config.LocalCfg.MaxDownload, config.LocalCfg.MaxUpload, OutEthRecvByteAvg, OutEthSendByteAvg, OutEthRecvSendUseRate)
			for iFace, ifInfo := range Net.IfiMap {
				ifRecv := FormatFloat64(ByteToBitM(ifInfo.RecvByteAvg))
				ifSend := FormatFloat64(ByteToBitM(ifInfo.SendByteAvg))
				ifRecvUseRate := FormatFloat64(ifRecv / OutIfMaxRecvBw)
				ifSendUseRate := FormatFloat64(ifSend / OutIfMaxSendBw)
				isInIp := Net.IsInIP(ifInfo)

				L.Infof("Iface details >>> iface: %v, isInIp: %v, ip: %v, ipMaxRecv: %vMbps, ipMaxSend: %vMbps, recv: %vMbps, send: %vMbps, recvUseRate: %v%%, sendUseRate:%v%%",
					iFace, isInIp, ifInfo.Ip, OutIfMaxRecvBw, OutIfMaxSendBw, ifRecv, ifSend, ifRecvUseRate, ifSendUseRate)
			}

			bandwidth := &Bandwidth{
				MachineId:   MachineID,
				Upload:      OutEthSendByteAvg,
				Download:    OutEthRecvByteAvg,
				MaxUpload:   float64(config.LocalCfg.MaxUpload),
				MaxDownload: float64(config.LocalCfg.MaxDownload),
				EthName:     StringsJoin(Net.IfiNames, ","),
			}

			go bandwidth.Report()
			time.Sleep(60 * time.Second)
		}
	}()
}
