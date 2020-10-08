package dis_redis

import (
	"database/sql"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

var (
	g errgroup.Group
	database1 *sqlx.DB
	pool1 *redis.Pool
	lockIF1 RedisLockIF
	timeOut=10
)



func testController( c *gin.Context) {
	//start
	timer1:=time.NewTimer(time.Second*time.Duration(timeOut))
	res:=make(chan int,1)
	defer func() {
		timer1.Stop()
	}()

	go func(res chan int) {
		u4 := uuid.New()
		s:=u4.String()
		e := lockIF1.Lock(s)
		if e!=nil {
			res<-2
			close(res)
			return
		}
		dealSkillLogic(database1,res,c.Query("port"))

		lockIF1.UnLock(s)
	}(res)

	select {
	case re:=<-res:
		if re == 1 {
			c.JSON(200,"skill success")
		}else{
			c.JSON(200,"skill error")
		}

		return
	case <-timer1.C:
		c.JSON(200,"timeout 10s")
		return


	}

}

func router02() http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())
	e.GET("/t1",testController)

	return e
}

func router01() http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())
	e.GET("/t1", testController)
	return e
}



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


func runserverDemo() {
	dns := "root:123456@tcp(127.0.0.1:3306)/skill"
	database, err := sqlx.Connect("mysql", dns)

	if err != nil {
		fmt.Print("mysql init error->",err)
		return
	}else{
		fmt.Println("init mysql success")
	}
	defer database.Close()
	database1=database

	pool:= getRedispool()
	pool1=pool
	lockIF, e := NewRedisLock(100, pool, "cisco", timeOut)
	if e!=nil {
		panic("参数异常")
	}
	lockIF1=lockIF

	server01 := &http.Server{
		Addr:         ":9091",
		Handler:      router01(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server02 := &http.Server{
		Addr:         ":9092",
		Handler:      router02(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	// 借助errgroup.Group或者自行开启两个goroutine分别启动两个服务
	g.Go(func() error {
		return server01.ListenAndServe()
	})

	g.Go(func() error {
		return server02.ListenAndServe()
	})
	fmt.Println("two server is running 9091 9092")
	if err := g.Wait(); err != nil {
		panic(err)
	}

}


func dealSkillLogic(db *sqlx.DB,res chan int,flg string)  {
	defer func() {
		close(res)
	}()
	tx,err:=db.Begin()
	if err!=nil {
		fmt.Println(err)
		tx.Rollback()
        defer func() {
			res<-2
		}()
		return
	}
	d:= struct {
		Num int `db:"num"`
	}{}
	err = db.Get(&d, "select num from good where id=? for update", 1)
	if err==sql.ErrNoRows {
		fmt.Println("no data")
		tx.Rollback()
		defer func() {
			res<-2
		}()
		return
	}
	if err!=nil {
		fmt.Println(err)
		tx.Rollback()
		defer func() {
			res<-2
		}()
		return
	}

	if d.Num>0 {
		_, err := tx.Exec("update good set num=num-1 where id=?", 1)
		if err != nil {
			// 失败回滚
			fmt.Println("err", err)
			tx.Rollback()
			defer func() {
				res<-2
			}()
			return
		}

		_, err = tx.Exec("insert into ord(order_no)values(?)", fmt.Sprintf("skillSuccess-port:%s",flg))
		if err != nil {
			// 失败回滚
			tx.Rollback()
			defer func() {
				res<-2
			}()
			return
		}

	}
	defer func() {
		res<-1
	}()
	tx.Commit()
	return




}
