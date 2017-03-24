package main

import (
	codis "git.pingxx.com/codis-client"
)

func main() {
	config := &codis.CodisConfig{
		CodisProxys: []codis.CodisProxy{
			codis.CodisProxy{
				Host:          "127.0.0.1",
				Port:          6701,
				PoolMaxIdle:   100,
				PoolMaxActive: 100,
				Password:      "",
			},
			codis.CodisProxy{
				Host:          "127.0.0.1",
				Port:          6702,
				PoolMaxIdle:   100,
				PoolMaxActive: 100,
				Password:      "",
			},
		},
	}
	client := codis.GetClient()
	client.SetConfig(config)
	client.Run()
}
