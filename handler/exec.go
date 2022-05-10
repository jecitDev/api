package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"strings"
	"time"

	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-redis/redis"
)

var ctx = context.Background()

//GoExec from db selection, return (*response, error)
func GoExec(p paramExec) (*response, error) {
	var err error
	x := paramDB{p.App, p.Entity, p.Token, p.Constr, p.Func, p.Attributes, p.Type, p.Data}
	var mData interface{}
	var respExec response
	respExec.Trid = Key() //time.Now().Format("20060102150405")
	var mmData map[string]interface{}
	redKey := p.Func
	if p.DbCache == "REDIS" {
		json.Unmarshal(p.Attributes, &mmData)
		if mmData["id"].(string) != "" {
			redKey = redKey + mmData["id"].(string)
		}
		if mmData["userID"].(string) != "" {
			redKey = redKey + mmData["userID"].(string)
		}
	}

	//Check data from db cache, before get data from primary db
	//---------------------------------------------------------
	if p.Crud == "R" {
		if p.DbCache == "REDIS" {
			// get data On db REDIS first
			Attributes := []byte(`{ "key": "` + redKey + `" }`)
			y := paramDB{p.App, p.Entity, p.Token, "cache", "GET", Attributes, "", nil}
			retRedis, err := GoExecredis(y)
			if err == nil { //no error
				if retRedis != nil { //have data
					json.Unmarshal(retRedis, &mData)
					respExec.Data = mData

					return &response{respExec.Trid, true, respExec.Message, respExec.Data}, err
				}
			}
			//no return, search on db.config
		} else if p.DbCache == "ES" {
			//get data from ES
			//no return, search on db.config
		}
	}

	// get data from primary db ============================================
	if p.DbType == "MSSQL" {
		retMSSQL, err := GoExecMSSQL(x)
		if err != nil {
			respExec.Message = "error executing query"
			return &response{respExec.Trid, false, respExec.Message, nil}, err
		}
		if retMSSQL == nil {
			respExec.Message = "no data can be display "
			return &response{respExec.Trid, false, respExec.Message, nil}, err
		}
		json.Unmarshal(retMSSQL, &mData)
		respExec.Data = mData

		// Caching data on DbCache
		if p.Crud == "R" {
			if p.DbCache == "REDIS" {
				// set value on redis for caching
				var mPar paramRedis
				mPar.Key = redKey
				mPar.Value = string(retMSSQL)
				Attributes, err := json.Marshal(mPar)

				y := paramDB{p.App, p.Entity, p.Token, "cache", "SET", Attributes, "", nil}
				retRedis, err := GoExecredis(y)
				check(err, "GoExecredis:Caching on Redis")
				if err != nil {
					fmt.Println(string(retRedis), err)
				}
			} else if p.DbCache == "ES" {
				// Write data On db ES
			}
		}

	} else if p.DbType == "POSTGRESQL" {
		retPostgreSQL, err := GoExecPostgreSQL(x)
		if err != nil {
			respExec.Message = "error executing query"
			return &response{respExec.Trid, false, respExec.Message, nil}, err
		}
		if retPostgreSQL == nil {
			respExec.Message = "no data can be display"
			return &response{respExec.Trid, false, respExec.Message, nil}, err
		}
		json.Unmarshal(retPostgreSQL, &mData)
		respExec.Data = mData

		// Caching data on DbCache
		if p.Crud == "R" {
			if p.DbCache == "REDIS" {
				// set value on redis for caching
				var mPar paramRedis
				mPar.Key = redKey
				mPar.Value = string(retPostgreSQL)
				Attributes, err := json.Marshal(mPar)

				y := paramDB{p.App, p.Entity, p.Token, "cache", "SET", Attributes, "", nil}
				retRedis, err := GoExecredis(y)
				check(err, "GoExecredis:Caching on Redis")
				if err != nil {
					fmt.Println(string(retRedis), err)
				}
			} else if p.DbCache == "ES" {
				// Write data On db ES
			}
		}

	} else {
		respExec.Message = "dbtype not recognized"
		return &response{respExec.Trid, false, respExec.Message, nil}, err
	}
	//======================================================================

	//On Command Successfully executed
	//--------------------------------
	//1. if Operation Create, also execute on db ES
	if p.Crud == "C" {
		if p.DbCache == "ES" {
			// Write data On db ES
		}
	}

	//2. if Operation Update Or Delete, also execute on db cache
	if (p.Crud == "U") || (p.Crud == "D") {
		if p.DbCache == "REDIS" {
			// Delete Cache On Redis
			Attributes := []byte(`{ "key": "` + redKey + `" }`)
			y := paramDB{p.App, p.Entity, p.Token, "cache", "DEL", Attributes, "", nil}
			retRedis, err := GoExecredis(y)
			check(err, "GoExecredis:Delete Cache")
			if err != nil { //no error
				fmt.Println(string(retRedis), err)
			}
		}
		if p.DbCache == "ES" {
			// Update On db ES
		}
	}

	return &response{respExec.Trid, true, respExec.Message, respExec.Data}, err
}

//========================================================================================================

//GoExecMSSQL file from db MSSQL
func GoExecMSSQL(p paramDB) ([]byte, error) {
	var err error
	flag.Parse()

	if p.Func == "GoGetConfig" {
		fmt.Println("-------------------------------------------------------")
	}
	sSource := "db system"
	if (p.Func == "GoGetConfig") || (p.Func == "GotransLog") {
		sSource = "db api"
	}
	fmt.Println("MSSQL connect " + sSource + ", exec " + p.Func)

	dsn := p.Constr
	db, err := sql.Open("mssql", dsn)
	check(err, "sqlOpen, Cannot connect:")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	err = db.PingContext(ctx)
	check(err, "dbPingContext, Error pinging db:")
	if err != nil {
		return nil, err
	}

	newData, err := execMSSQL(db, p)
	if err != nil {
		return nil, err
	}

	return newData, err
}

//GoExecPostgreSQL file from db PostGreSQL
func GoExecPostgreSQL(p paramDB) ([]byte, error) {
	var err error

	newData := []byte(`[ { "Message": "GoExecPostgreSQL masih belom ada kodingannya", "Result": 0 } ]`)
	return newData, err
}

//GoExecredis file from db redis
func GoExecredis(p paramDB) ([]byte, error) {
	var err error
	var newData []byte
	var mPar paramRedis
	json.Unmarshal(p.Attributes, &mPar)
	if mPar.Key == "" {
		newData = []byte(`[ { "Message": "redis need key attribute.", "Result": 0 } ]`)
		return newData, err
	}

	fmt.Println("redis " + p.Func + " " + p.Constr + ", key: " + mPar.Key)
	redisClient, err := initialize()
	if err != nil {
		newData = []byte(`[ { "Message": "can not connect to redis db.", "Result": 0 } ]`)
		return newData, err
	}

	if p.Func == "GET" {
		retRedis, err := redisClient.getKey(mPar.Key)
		if err != nil {
			return nil, err
		}
		str := fmt.Sprintf("%v", retRedis)

		if isJSON(str) == true {
			newData = []byte(str)
		} else {
			// return JSON format
			if p.Constr == "cache" {
				newData = []byte(str)
				//newData = []byte(`[ { "Data": "` + str + `", "Message": "data cache is not JSON format.", "Result": 1 } ]`)
			} else if p.Constr == "config" {
				newData = []byte(`{ "Data": "` + str + `" }`)
			}
		}

	} else if p.Func == "SET" {
		if mPar.Value == "" {
			newData = []byte(`[ { "Message": "redis need value attribute.", "Result": 0 } ]`)
			return newData, err
		}

		expiration := time.Minute * 15 //expiration logic; mPar.Expiration LIAR banget
		err := redisClient.setKey(mPar.Key, mPar.Value, expiration)
		check(err, "redisClient.setKey:")
		if err != nil {
			return nil, err
		}

		newData = []byte(`[ { "Message": "data saved for ` + fmt.Sprint(expiration) + `.", "Result": 1 } ]`)

	} else if p.Func == "DEL" {

		err := redisClient.delKey(mPar.Key)
		check(err, "redisClient.delKey:")
		if err != nil {
			return nil, err
		}
		newData = []byte(`[ { "Message": "data has been deleted.", "Result": 1 } ]`)

	} else {
		newData = []byte(`[ { "Message": "redis Func(` + p.Func + `) not recognation.", "Result": 0 } ]`)
		return newData, err
	}

	return newData, err
}

//GoExecES file from db Elastic Search
func GoExecES(p paramDB) ([]byte, error) {
	var err error

	newData := []byte(`[ { "Message": "GoExecES masih belom ada kodingannya", "Result": 0 } ]`)
	return newData, err
}

//========================================================================================================

//This is func For GoExecMSSQL
func execMSSQL(db *sql.DB, p paramDB) ([]byte, error) {
	var err error
	var mParSP map[string]interface{}
	json.Unmarshal(p.Attributes, &mParSP)

	cmd := " "
	if p.Type == "setbulk" {
		cmd = cmd + " Select Getdate() as U "
	} else {
		cmd = cmd + " Exec " + p.Func
		cmd = cmd + " @app='" + p.App + "',"
		cmd = cmd + " @entity='" + p.Entity + "',"
		cmd = cmd + " @Token='" + p.Token + "',"
		for key, value := range mParSP {
			cmd = cmd + " @" + key + "='" + strings.ReplaceAll(value.(string), "'", "''") + "',"
		}
		cmd = cmd[:len(cmd)-1]
	}

	rows, err := db.Query(cmd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	check(err, "execMSSQL, Error rows.Columns():")
	if err != nil {
		return nil, err
	}
	if cols == nil {
		fmt.Println("execMSSQL, Error nil cols: ")
		return nil, err
	}
	if len(cols) == 0 {
		return nil, err
	}

	// convert (rows *sql.Rows) to (JSON []byte)
	// =========================================
	values := make([]interface{}, len(cols))
	results := make(map[string]interface{})
	for i := 0; i < len(cols); i++ {
		values[i] = new(interface{})
	}

	var data []byte
	c := 0

	for rows.Next() {
		//for i := range values {
		//	values[i] = new(interface{})
		//}

		err = rows.Scan(values...)
		check(err, "execMSSQL, Error rows.Scan(values...):")
		if err != nil {
			continue
		}

		for i, v := range values {
			var iTempSw interface{}
			switch z := (*(v.(*interface{}))).(type) {
			case nil:
				iTempSw = nil
			case float64:
				iTempSw = Round(z, 5)
			case []byte:
				iTempSw = string(z)
			case bool:
				if z {
					iTempSw = 1
				} else {
					iTempSw = 0
				}
			case time.Time:
				tTemp := []rune(z.Format("2006-01-02 15:04:05"))
				if string(tTemp[11:19]) == "00:00:00" { //date
					iTempSw = string(tTemp[0:10])
				} else if string(tTemp[0:10]) == "0001-01-01" { //time
					iTempSw = string(tTemp[11:19])
				} else { //datetime
					iTempSw = string(tTemp)
				}
			default:
				iTempSw = z
			}
			results[cols[i]] = iTempSw
		}

		b, err := json.Marshal(results)
		check(err, "execMSSQL, Error json.Marshal(results):  ")
		if err != nil {
			continue
		}

		if c > 0 {
			data = append(data, ","...)
		}
		data = append(data, string(b)...)

		c++
	}

	var arData []byte
	if p.Func == "GoGetConfig" {
		arData = data
	} else if p.Func == "GotransLog" {
		arData = data
	} else {
		if string(data) == "" {
			fmt.Println(cmd)
			arData = data
		} else {
			arData = append(arData, "["...)
			arData = append(arData, data...)
			arData = append(arData, "]"...)
		}
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return arData, err
}

//transLog : write log
func transLog(p paramLog) {
	Constr := varConfig("dbapi")
	sQ := fmt.Sprint(p.Quota)
	sEt := fmt.Sprint(p.ExecTime)
	bS := fmt.Sprint(p.Success)
	Attributes := []byte(`{ "function": "` + p.Function + `", "quota": "` + sQ + `", "execTime": "` + sEt + `", "success": "` + bS + `", "status": "` + p.Status + `", "message": "` + p.Message + `", "userID": "` + p.UserID + `" }`)
	x := paramDB{p.App, p.Entity, p.Token, Constr, "GotransLog", Attributes, "", nil}
	retMSSQL, err := GoExecMSSQL(x)
	check(err, "transLog:")
	var mData map[string]interface{}
	json.Unmarshal(retMSSQL, &mData)
	if mData["Result"].(float64) == 0 {
		fmt.Println(mData["Message"].(string))
	}

	return
}

//GetClient get the redis client
func initialize() (*redisClient, error) {
	var err error
	address := varConfig("dbredis")
	pwd := varConfig("redispwd")

	rdb := redis.NewClient(&redis.Options{
		Addr:       address,
		Password:   pwd,
		DB:         0,
		MaxRetries: 3,
	})

	if err = rdb.Ping().Err(); err != nil {
		//check(err, "initialize redis:")
		return client, err
	}

	client.c = rdb
	//fmt.Println(client.c)
	return client, err
}

//GetKey get key redis
func (client *redisClient) getKey(key string) (interface{}, error) {
	val, err := client.c.Get(key).Result()
	if err != nil || err == redis.Nil {
		return nil, err
	}
	return val, nil
}

//SetKey set key redis
func (client *redisClient) setKey(key string, value interface{}, expiration time.Duration) error {
	err := client.c.Set(key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

//delKey delete key redis
func (client *redisClient) delKey(key string) error {
	err := client.c.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}
