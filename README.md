# dis-redis 基于redis的分布式锁
```cassandraql
lockIF, e := NewRedisLock(100, pool, "cisco", timeout)
	if e!=nil {
		panic("参数异常")
	}


e := lockIF.Lock(v)
	if e!=nil {
		return
	}

//业务
lockIF.Unlock()

```
