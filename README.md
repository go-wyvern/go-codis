## Demo示例

```
package main

import (
	"time"

	"fmt"

	codis "git.pingxx.com/codis-client"
)

func main() {
	config := &codis.CodisConfig{
		TickDuration: 5 * time.Second,
		CodisProxys: []codis.CodisProxy{
			codis.CodisProxy{
				Host:          "127.0.0.1",
				Port:          6379,
				PoolMaxIdle:   100,
				PoolMaxActive: 100,
				Password:      "",
			},
			codis.CodisProxy{
				Host:          "127.0.0.1",
				Port:          6378,
				PoolMaxIdle:   100,
				PoolMaxActive: 100,
				Password:      "",
			},
		},
	}
	client := codis.GetClient(config)
	go client.Run()
	for {
		time.Sleep(5 * time.Second)
		conn := codis.Get()
		if conn == nil {
			fmt.Println("没有可用的链接")
			continue
		}
		_, err := conn.Do("set", "a", "1")
		if err != nil {
			fmt.Println("SET Error", err)
			continue
		}
		fmt.Println("SET OK")
	}
}

```

- Ha检测
- 自动切换