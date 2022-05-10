package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unsafe"
)

//GoPostData is func to post data to db with return success or not
func GoPostData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	var tStart = time.Now()

	// get value from param url & body
	//====================================================================================================
	pathParams, err := getParam(w, r)
	if err != nil {
		fmt.Println("✨GoPostData, Error getParam: ", err.Error())
		return
	}
	if pathParams.Success == false {
		fmt.Println("✨GoPostData, ", pathParams.Message)
		return
	}

	// get function config from api datastore config
	//====================================================================================================
	responseConfig, err := GetConfig(paramConfig{pathParams.App, pathParams.Entity, pathParams.Token, pathParams.Function, pathParams.Attributes})
	check(err, "GoPostData, Error GoGetConfig:")
	if err != nil {
		http.Error(w, `{ "success": false, "message": "can not connect", "data": [ { "Message": "`+err.Error()+`", "Result": 0 } ] }`, http.StatusInternalServerError)
		return
	}
	if responseConfig.Success == false {
		fmt.Println("✨GoPostData, ", responseConfig.Message)
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		w.Write([]byte(`{ "success": false, "message": "` + responseConfig.DbType + `", "data": [ { "Message": "` + responseConfig.Message + `", "Result": 0 } ] }`))
		transLog(paramLog{pathParams.App, pathParams.Entity, pathParams.Token, pathParams.Function, false, unsafe.Sizeof(responseConfig), time.Since(tStart).Milliseconds(), "StatusNonAuthoritativeInfo", responseConfig.Message, pathParams.UserID})
		return
	}

	// execute api request
	//====================================================================================================
	returndata, err := GoExec(paramExec{pathParams.App, pathParams.Entity, pathParams.Token, responseConfig.DbType, responseConfig.Constr, pathParams.Function, pathParams.Attributes, pathParams.Type, pathParams.Data, responseConfig.DbCache, responseConfig.Crud})
	check(err, "GoPostData, Error GoExec:")
	if err != nil {
		s := strings.ReplaceAll(err.Error(), "stored procedure", "function")
		s = strings.ReplaceAll(s, "'", "")
		http.Error(w, `{ "success": false, "message": "error executing query", "data": [ { "Message": "`+s+`", "Result": 0 } ] }`, http.StatusInternalServerError)
		transLog(paramLog{pathParams.App, pathParams.Entity, pathParams.Token, pathParams.Function, false, unsafe.Sizeof(pathParams.Attributes), time.Since(tStart).Milliseconds(), "StatusInternalServerError", s, pathParams.UserID})
		return
	}
	if returndata == nil {
		fmt.Println("✨GoPostData, returndata = nil ")
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`{ "success": false, "message": "no data can be display", "data": [ { "Message": "` + err.Error() + `", "Result": 0 } ] }`))
		transLog(paramLog{pathParams.App, pathParams.Entity, pathParams.Token, pathParams.Function, false, unsafe.Sizeof(pathParams.Attributes), time.Since(tStart).Milliseconds(), "StatusNoContent", "no data can be display", pathParams.UserID})
		return
	}

	jData, err := json.Marshal(returndata)
	check(err, "GoPostData, Error json.Marshal(returndata):")
	if err != nil {
		http.Error(w, `{ "success": false, "message": "error marshalling data", "data": [ { "Message": "`+err.Error()+`", "Result": 0 } ] }`, http.StatusInternalServerError)
		transLog(paramLog{pathParams.App, pathParams.Entity, pathParams.Token, pathParams.Function, false, unsafe.Sizeof(jData), time.Since(tStart).Milliseconds(), "StatusInternalServerError", "error marshalling data", pathParams.UserID})
		return
	}

	// return value and logging for the request
	// Success Logging turning off, until byte access validate activated; since v2.1
	//====================================================================================================
	w.WriteHeader(http.StatusOK)
	w.Write(jData)
	//transLog(paramLog{pathParams.App, pathParams.Entity, pathParams.Token, pathParams.Function, true, unsafe.Sizeof(jData), time.Since(tStart).Milliseconds(), "StatusOK", "", pathParams.UserID})
	return
}
