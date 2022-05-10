package handler

import (
	"encoding/json"
	"fmt"
)

// GetConfig ... , get configuration for this app parameter
func GetConfig(p paramConfig) (*responseConfig, error) {
	var err error
	var respCon responseConfig
	var mData map[string]interface{}

	//get config "dbType & conStr" from dbapi configuration
	//====================================================================================================
	//1. Validate function app & Token
	//2. validate reQuirement body param for this function (this point is not ready)
	Constr := varConfig("dbapi")
	Attributes := []byte(`{ "function": "` + p.Function + `" }`)
	x := paramDB{p.App, p.Entity, p.Token, Constr, "GoGetConfig", Attributes, "", nil}
	retMSSQL, err := GoExecMSSQL(x)
	if err != nil {
		return &responseConfig{false, "error connect to api db", "can not connect", "", "", ""}, err
	}
	if retMSSQL == nil {
		return &responseConfig{false, "no configuration for your app", "can not connect", "", "", ""}, err
	}

	json.Unmarshal(retMSSQL, &mData)
	if mData["Result"].(float64) == 0 {
		respCon.Message = mData["Message"].(string)
		return &responseConfig{false, respCon.Message, "Validate Failed:", "", "", ""}, err
	}
	if mData["dbType"].(string) == "" {
		respCon.Message = mData["Message"].(string)
		return &responseConfig{false, respCon.Message, "Validate Failed:", "", "", ""}, err
	}
	if mData["conStr"].(string) == "" {
		respCon.Message = mData["Message"].(string)
		return &responseConfig{false, respCon.Message, "Validate Failed:", "", "", ""}, err
	}
	respCon.DbType = mData["dbType"].(string)
	respCon.Constr = mData["conStr"].(string)

	//get config "DbCache & crud" app from redis configuration
	//====================================================================================================
	id := ""
	var mDataAtt map[string]interface{}
	json.Unmarshal(p.Attributes, &mDataAtt)
	if mDataAtt["id"] != nil {
		if mDataAtt["id"].(string) != "" {
			id = mDataAtt["id"].(string)
		}
	}

	respCon.DbCache = ""
	respCon.Crud = ""
	Attributes = []byte(`{ "key": "` + p.Function + id + `" }`)
	// y := paramDB{p.App, p.Entity, p.Token, "config", "GET", Attributes, "", nil}
	// retRedis, err := GoExecredis(y)
	// if err != nil {
	// 	fmt.Println("🦄 "+string(Attributes), err)
	// 	err = nil // error on get config redis do not stop the process
	// } else {
	// 	if retRedis != nil { //have data
	// 		// format data config:  { "dbCache": "REDIS", "crud": "R" }
	// 		var mDataX map[string]interface{}
	// 		json.Unmarshal(retRedis, &mDataX)
	// 		if mDataX["dbCache"] != nil {
	// 			if mDataX["dbCache"].(string) != "" {
	// 				respCon.DbCache = mDataX["dbCache"].(string)
	// 			}
	// 		}
	// 		if mDataX["crud"] != nil {
	// 			if mDataX["crud"].(string) != "" {
	// 				respCon.Crud = mDataX["crud"].(string)
	// 			}
	// 		}
	fmt.Println(respCon.DbCache, respCon.Crud)
	// 	}
	// }

	return &responseConfig{true, "", respCon.DbType, respCon.Constr, respCon.DbCache, respCon.Crud}, err
}

//varConfig: declaration module
func varConfig(p string) string {
	var rP string

	if p == "dbapi" {
		//rP = "server=.;user id=sa;Password=4dm1n.db;database=api"
		//rP = "server=172.16.20.30;user id=apigo;Password=4dm1n.Api;database=api"
		rP = "server=172.16.80.167;port=51433;user id=apigo;Password=4dm1n.Api;database=api"
	} else if p == "dbredis" {
		rP = "localhost:6300"
		//rP = "172.16.80.167:6300"
	} else if p == "redispwd" {
		rP = "ganteng banget"
	}

	return rP
}
