package rss

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Dcarbon/go-shared/libs/dbutils"
	"github.com/Dcarbon/go-shared/libs/ievent"
	"github.com/Dcarbon/go-shared/libs/rabbit"
	"github.com/Dcarbon/go-shared/libs/utils"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var dbUrl = utils.StringEnv("DB_URL", "postgres://admin:hellosecret@localhost/iot_op")
var amqpUrl = utils.StringEnv("AMQP_URL", "amqp://rbuser:hellosecret@localhost")
var redisUrl = utils.StringEnv("REDIS_URL", "redis://localhost:6379")

var mut = &sync.Mutex{}

var singDB *gorm.DB
var redisClient *redis.Client
var singRabbit rabbit.IConnection

func SetUrl(dbUrlConn, redisUrlConn string) {
	if dbUrlConn != "" {
		dbUrl = dbUrlConn
		GetDB()
	}
	if redisUrlConn != "" {
		redisUrl = redisUrlConn
		GetRedis()
	}
}

func GetDB() *gorm.DB {
	if nil == singDB {
		mut.Lock()
		defer mut.Unlock()
		var err error
		if nil == singDB {
			dbutils.CreateDB(dbUrl)
			singDB, err = dbutils.NewDB(dbUrl)
			if nil != err {
				panic(err)
			}

			// singDB.Logger = logger.New(
			// 	log.New(os.Stdout, "\r\n", log.LstdFlags),
			// 	logger.Config{
			// 		LogLevel: logger.Info,
			// 	},
			// )
		}
	}
	return singDB
}

func GetRedis() *redis.Client {
	if nil == redisClient {
		mut.Lock()
		defer mut.Unlock()

		if nil == redisClient {
			opt, err := redis.ParseURL(redisUrl)
			if nil != err {
				panic(fmt.Errorf("parse redis url[%s] error: %s", redisUrl, err.Error()))
			}
			redisClient = redis.NewClient(opt)

			_, err = redisClient.Ping(context.TODO()).Result()
			if nil != err {
				panic(errors.New("ping to redis error: " + err.Error()))
			}
		}
	}
	return redisClient
}

func GetRabbitMQ() rabbit.IConnection {
	if singRabbit == nil {
		mut.Lock()
		defer mut.Unlock()

		if singRabbit == nil {
			var err error
			singRabbit, err = rabbit.Dial(amqpUrl)
			if nil != err {
				panic(errors.New("Connect to rabbit mq error: " + err.Error()))
			}
		}
	}
	return singRabbit
}

func GetRabbitPusher() ievent.IPublisher {
	var rbConn = GetRabbitMQ()

	rbChan, err := rbConn.Channel()
	utils.PanicError("", err)

	var pusher = ievent.NewDirectPusher(rbChan)
	return pusher
}
