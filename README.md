# dis-redis 基于redis的分布式锁

### 使用方法
```cassandraql
import  "github.com/liuhangkaixcode/dis-redis"
lockIF, e := dis_redis.NewRedisLock(100, pool, "cisco", timeout)
	if e!=nil {
		panic("参数异常")
	}

//v 是 setnx 设置的value
e := lockIF.Lock(v)
	if e!=nil {
		return
	}

....这是处理业务

lockIF.Unlock(v)

```


### demo1 TestNewRedisLock 模拟抢票过程
------------------------------------------------------------------------
###  demo2  TestGinServer  模拟分布式秒杀过程

ab.exe -c 100 -n 1000  http://127.0.0.1:9091/t1?port=9091
### 
ab.exe -c 100 -n 1000  http://127.0.0.1:9092/t1?port=9092

下面是数据库的结构
```cassandraql
mysql> select * from good;
+----+---------+--------+-----+
| id | good_id | name   | num |
+----+---------+--------+-----+
|  1 |       1 | orange | 111 |
+----+---------+--------+-----+
```

```cassandraql
mysql> desc ord;
+----------+------------------+------+-----+---------+----------------+
| Field    | Type             | Null | Key | Default | Extra          |
+----------+------------------+------+-----+---------+----------------+
| id       | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
| order_no | varchar(100)     | YES  |     | NULL    |                |
+----------+------------------+------+-----+---------+----------------+
```
