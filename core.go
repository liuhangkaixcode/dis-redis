package dis_redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

type redisLock struct {
	name string
	timeout int
	l sync.Mutex
	pool *redis.Pool
	trylocknum int
}
type RedisLockIF interface {
	Lock(v string) error
	UnLock(v string)error
}
//trylocknum尝试加锁的次数
//pool redis线程池
//key  setnx的key
//timeout 加锁成功后，业务最大处理时间
func NewRedisLock(trylocknum int, pool *redis.Pool, key string,timeout int) (RedisLockIF,error) {

	if pool==nil {
		return nil,fmt.Errorf("redis.pool is nil")
	}
	if len(key) == 0 {
		return nil,fmt.Errorf("key is nil")
	}
	if timeout == 0 {
		return nil,fmt.Errorf("timeout is nil")
	}
	if trylocknum<10 {
		trylocknum=100
	}
	a:=new(redisLock)
	a.name=key
	a.pool=pool
	a.trylocknum=trylocknum
	a.timeout=timeout
	return a,nil
}

func (p *redisLock)Lock(v string) error{
	conn:=p.pool.Get()
	p.l.Lock()
	defer func() {
		conn.Close()
		p.l.Unlock()
	}()
	for i:=0;i<p.trylocknum;i++{
		_, err1 := redis.String(conn.Do("SET", p.name, v, "EX", p.timeout, "NX"))
		if err1 != nil {

			if err1==redis.ErrNil { //
				ttl1, e := redis.Int(conn.Do("TTL", p.name))
				if e!=nil {
					continue
				}
				if ttl1>0 {
					time.Sleep(time.Duration(ttl1*1000/100)*time.Millisecond)

				}
				if ttl1 == -1 {
					conn.Do("Del", p.name)
				}
				continue
			}
			return err1

		}

		if i>=p.trylocknum-2 {
			return fmt.Errorf("尝试加锁，没有加锁成功")
		}
		//get locker 获取到锁了
		return nil
	}
	return nil

}

func (p *redisLock)UnLock( v string)  error{
	conn:=p.pool.Get()
	defer conn.Close()
	vt, err:= redis.String(conn.Do("get", p.name))
	if vt == v{
		conn.Do("Del", p.name)
	}

	return err
}

