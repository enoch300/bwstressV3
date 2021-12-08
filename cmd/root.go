/*
* @Author: wangqilong
* @Description:
* @File: root
* @Date: 2021/9/17 3:16 下午
 */

package cmd

import (
	"github.com/spf13/cobra"
	"ipaas_bwstress/server"
	"ipaas_bwstress/util/config"
	"ipaas_bwstress/util/log"
	"os"
)

var logpath string
var rootCmd = &cobra.Command{
	Use:     "bwstress",
	Short:   "带宽补量程序",
	Version: "3.2.0",
	Run: func(cmd *cobra.Command, args []string) {
		log.New(logpath, "bwstress", 3)
		server.Run()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&config.API, "api", "a", "https://ipaas.paigod.work/api/v1/bwstress", "ipaas center api")
	rootCmd.Flags().StringVarP(&logpath, "log-path", "l", "/ipaas/bwstress/logs", "log path")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
