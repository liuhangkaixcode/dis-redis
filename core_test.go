package dis_redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
	"sync"
	"testing"
	"time"
)

var (
	ws sync.WaitGroup
	ticketSum=15
	timeout=10
)
//初始化一个redispool
func getRedispool()  *redis.Pool{
	pool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   20000,
		IdleTimeout: 10 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
	}
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("ping")
	if err != nil {
		panic("redis is not start........\n")
	}else{
		fmt.Println("redis inint success....")
	}
	return pool
}

func TestNewRedisLock(t *testing.T) {
      pool:= getRedispool()
	lockIF, e := NewRedisLock(100, pool, "cisco", timeout)
	if e!=nil {
		panic("参数异常")
	}


	for i:=0;i<200;i++{
		ws.Add(1)
		go ticketgrabbing(lockIF,pool)
	}
	ws.Wait()

}
func ticketgrabbing(lockIF RedisLockIF,pool *redis.Pool)  {
	u4 := uuid.New()
	v:=u4.String()
	e := lockIF.Lock(v)
	if e!=nil {
		ws.Done()
		return
	}
	res:=make(chan int ,1)
	timer:=time.NewTicker(time.Duration(timeout)*time.Second)
	defer func() {
		lockIF.UnLock(v)
		timer.Stop()
		ws.Done()
	}()

	go func(res chan int) {
		if ticketSum>0 {
			ticketSum=ticketSum-1
			res<-1
		}else{
			res<-2
		}


	}(res)

	select {
	   case t:=<-res:
		   if t==1 {
			   fmt.Printf("抢到票了，还剩余%d\n",ticketSum)
		   }
		   return
	case <-timer.C:
		fmt.Printf("超时了%d秒\n",timeout)
		return


	}


}
