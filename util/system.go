/*
* @Author: wangqilong
* @Description:
* @File: system
* @Date: 2021/10/11 12:51 下午
 */

package util

import (
	"io/ioutil"
	"log"
	"strings"
)

var MachineID string

func GetMachineId() {
	machineId, err := ioutil.ReadFile("/etc/machine-id")
	if err != nil {
		log.Fatalf("Get machineID: %v\n", err.Error())
	}
	MachineID = strings.TrimSpace(string(machineId))
}
