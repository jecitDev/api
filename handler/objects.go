package handler

import (
	"time"

	"github.com/go-redis/redis"
)

var (
	client = &redisClient{}
)

//type redisClient, declaration
type redisClient struct {
	c *redis.Client
}

//type response, return value "Data" is JSON format.
type response struct {
	Trid    string      `json:"trid"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type resVal []response

//type paramRedis, return data from redis
type paramRedis struct {
	Key        string        `json:"key"`
	Value      interface{}   `json:"value"`
	Expiration time.Duration `json:"expiration"`
}
type parRed []paramRedis

//type responseConfig, return value configuration
type responseConfig struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	DbType  string `json:"dbType"`
	Constr  string `json:"constr"`
	DbCache string `json:"dbCache"`
	Crud    string `json:"crud"`
}
type resCon []responseConfig

//type responseParam, return value param url & body
type responseParam struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	App        string `json:"app"`
	Entity     string `json:"entity"`
	Token      string `json:"token"`
	Function   string `json:"function"`
	Attributes []byte `json:"attributes"`
	Type       string `json:"type"`
	Data       []byte `json:"data"`
	UserID     string `json:"userID"`
}
type resPar []responseParam

//type paramConfig, is object parameter for get Config
type paramConfig struct {
	App        string `json:"app"`
	Entity     string `json:"entity"`
	Token      string `json:"token"`
	Function   string `json:"function"`
	Attributes []byte `json:"attributes"`
}
type parConfig []paramConfig

//type paramExec, is object parameter for execute data
type paramExec struct {
	App        string `json:"app"`
	Entity     string `json:"entity"`
	Token      string `json:"token"`
	DbType     string `json:"dbType"`
	Constr     string `json:"constr"`
	Func       string `json:"func"`
	Attributes []byte `json:"attributes"`
	Type       string `json:"type"`
	Data       []byte `json:"data"`
	DbCache    string `json:"dbCache"`
	Crud       string `json:"crud"`
}
type parExec []paramExec

//type paramDB, is object parameter for db execute
type paramDB struct {
	App        string `json:"app"`
	Entity     string `json:"entity"`
	Token      string `json:"token"`
	Constr     string `json:"constr"`
	Func       string `json:"func"`
	Attributes []byte `json:"attributes"`
	Type       string `json:"type"`
	Data       []byte `json:"data"`
}
type parDB []paramDB

//type paramLog, is object parameter for api logging
type paramLog struct {
	App      string  `json:"app"`
	Entity   string  `json:"entity"`
	Token    string  `json:"token"`
	Function string  `json:"function"`
	Success  bool    `json:"success"`
	Quota    uintptr `json:"quota"`    // in Byte
	ExecTime int64   `json:"execTime"` // in milisecond
	Status   string  `json:"status"`
	Message  string  `json:"message"`
	UserID   string  `json:"userID"`
}
type parLog []paramLog
