/* *
 * Ping++ Codis SDK
 * 使用说明：
 *    import (
 *         "git.pingxx.com/api/sdk-codis/go-codis"
 *	  )
 *    func main(){
 *	 	config := &codis.CodisConfig{
 *		    CodisProxys: []codis.CodisProxy{
 *				codis.CodisProxy{
 *					Host:          "127.0.0.1",
 *					Port:          6701,
 *					PoolMaxIdle:   100,
 *					PoolMaxActive: 100,
 *					Password:      "",
 *				},
 *				codis.CodisProxy{
 *					Host:          "127.0.0.1",
 *					Port:          6702,
 *					PoolMaxIdle:   100,
 *					PoolMaxActive: 100,
 *					Password:      "",
 *				},
 *			},
 *	  	}
 *		client := codis.GetClient()
 *		client.SetConfig(config)
 *		client.Run()
 *    }
 *
 *
 *
 *
 */

package codis

import (
	"container/list"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

type CodisClient struct {
	C            []*redis.Pool
	Config       *CodisConfig
	mu           sync.Mutex
	okList       list.List
	errList      list.List
	tickDuration time.Duration
	ticker       *time.Ticker
	errchan      chan *redis.Pool
	okchan       chan *redis.Pool
	exitchan     chan bool
}

var defaultCodisClient *CodisClient

type CodisProxy struct {
	PoolMaxIdle   int
	PoolMaxActive int
	Host          string
	Port          int
	Password      string
}

type CodisConfig struct {
	CodisProxys []CodisProxy
}

func GetClient() *CodisClient {
	return defaultCodisClient
}

func (r CodisProxy) Address() string {
	return r.Host + ":" + strconv.Itoa(r.Port)
}

func (client *CodisClient) SetConfig(config *CodisConfig) {
	client.Config = config
}

func clientMonitor() {
	for {
		select {
		case <-defaultCodisClient.exitchan:
			goto exits
		case rp := <-defaultCodisClient.okchan:
			defaultCodisClient.okList.PushBack(rp)
		case rp := <-defaultCodisClient.errchan:
			defaultCodisClient.errList.PushBack(rp)
		case <-defaultCodisClient.ticker.C:
			clientCheckOk()
			clientCheckErr()
		}

	}
exits:
}

func (client *CodisClient) Run() {
	ClientInit(client.Config)
	clientMonitor()
}

func clientCheckOk() {
	for i, n := 0, defaultCodisClient.okList.Len(); i < n; i++ {
		e := defaultCodisClient.okList.Front()
		if e == nil {
			break
		}
		rp := e.Value.(*redis.Pool)
		test := rp.TestOnBorrow

		if c := rp.Get(); test == nil && test(c, time.Now()) != nil {
			defaultCodisClient.okList.Remove(e)
			defaultCodisClient.errList.PushBack(rp)
		} else {
			c.Close()
		}
	}
}

func clientCheckErr() {
	for i, n := 0, defaultCodisClient.errList.Len(); i < n; i++ {
		e := defaultCodisClient.errList.Front()
		if e == nil {
			break
		}
		rp := e.Value.(*redis.Pool)
		test := rp.TestOnBorrow

		if c := rp.Get(); test == nil && test(c, time.Now()) == nil {
			defaultCodisClient.errList.Remove(e)
			defaultCodisClient.okList.PushBack(rp)
		} else {
			c.Close()
		}
	}
}

func ClientInit(c *CodisConfig) {
	defaultCodisClient = new(CodisClient)
	defaultCodisClient.ticker = time.NewTicker(defaultCodisClient.tickDuration)
	defaultCodisClient.errchan = make(chan *redis.Pool, 16)
	defaultCodisClient.okchan = make(chan *redis.Pool, 16)
	defaultCodisClient.exitchan = make(chan bool)
	for _, proxy := range c.CodisProxys {
		rp := InitRedisPool(proxy)
		defaultCodisClient.C = append(defaultCodisClient.C, rp)
		if err := rp.TestOnBorrow(rp.Get(), time.Now()); err != nil {
			defaultCodisClient.errList.PushFront(rp)
		} else {
			defaultCodisClient.okList.PushFront(rp)
		}
	}
}

func InitRedisPool(proxy CodisProxy) *redis.Pool {
	var rp *redis.Pool
	rp = &redis.Pool{
		// 最大空闲连接
		MaxIdle: proxy.PoolMaxIdle,
		// 最大活跃连接
		MaxActive: proxy.PoolMaxActive,
		// 超时时间
		IdleTimeout: 30 * time.Second,
		// 连接创建函数
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", proxy.Address(), redis.DialConnectTimeout(10*time.Second))
			if err != nil {
				return nil, err
			}
			if proxy.Password != "" {
				if _, err := conn.Do("AUTH", proxy.Password); err != nil {
					conn.Close()
					return nil, err
				}
			}
			return conn, err
		},
		// 连接测试函数
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return rp
}

func Get() redis.Conn {
	e := defaultCodisClient.okList.Front()
	if e == nil {
		return nil
	}
	rp := e.Value.(*redis.Pool)
	return rp.Get()
}
