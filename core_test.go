package dis_redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
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

//模拟并发抢票的过程
func TestNewRedisLock(t *testing.T) {
	//redis pool初始化
	pool:= getRedispool()
	lockIF, e := NewRedisLock(100, pool, "cisco", timeout)
	if e!=nil {
		panic("参数异常")
	}

	//模拟20000 并发
	for i:=0;i<20000;i++{
		ws.Add(1)
		go ticketgrabbing(lockIF,pool)
	}
	ws.Wait()

}
//抢票逻辑
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


//模拟分布式秒杀过程
func TestGinServer(t *testing.T)  {
	runserverDemo()
}
