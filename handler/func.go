package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

//getParam : get param url & body
func getParam(w http.ResponseWriter, r *http.Request) (*responseParam, error) {
	pathParams := mux.Vars(r)
	var respPar responseParam
	var err error

	// get value from param url
	//=========================
	if val, ok := pathParams["app"]; ok {
		if val == "" {
			respPar.Message = `{ "success": false, "message": "need an app parameter", "data": [ { "Message": "error", "Result": 0 } ] }`
			http.Error(w, respPar.Message, http.StatusBadRequest)
			return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
		}
		respPar.App = val
	}
	if val, ok := pathParams["entity"]; ok {
		if val == "" {
			respPar.Message = `{ "success": false, "message": "need an entity parameter", "data": [ { "Message": "error", "Result": 0 } ] }`
			http.Error(w, respPar.Message, http.StatusBadRequest)
			return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
		}
		respPar.Entity = val
	}

	// get value from param body
	//==========================
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("🐶getParam, Error ioutil.ReadAll(r.Body): ", err.Error())
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "` + err.Error() + `", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	if string(reqBody) == "" {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "please re-check your JSON body parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	if strings.ReplaceAll(strings.ReplaceAll(string(reqBody), "{", ""), "}", "") == "" {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "please re-check your JSON body parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	if isJSON(string(reqBody)) == false {
		fmt.Println(string(reqBody))
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "please re-check your JSON body Format", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}

	var mBody map[string]interface{}
	json.Unmarshal(reqBody, &mBody)
	if mBody["token"] == nil {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "need an Token parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	if mBody["token"] == "" {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "need an Token parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	respPar.Token = mBody["token"].(string)

	if mBody["function"] == nil {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "need an Function parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	if mBody["function"] == "" {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "need an Function parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	respPar.Function = mBody["function"].(string)

	if mBody["attributes"] == nil {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "need an attributes parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	mAttributes := mBody["attributes"].(map[string]interface{})
	jmAttributes, err := json.Marshal(mAttributes)
	check(err, "getParam, Error json.Marshal(mAttributes):")
	if err != nil {
		respPar.Message = `{ "success": false, "message": "error in attributes body parameter", "data": [ { "Message": "` + err.Error() + `", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	if strings.ReplaceAll(strings.ReplaceAll(string(jmAttributes), "{", ""), "}", "") == "" {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "attributes parameter ca no be blank", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	respPar.Attributes = jmAttributes
	if mAttributes["userID"] == nil {
		respPar.Message = `{ "success": false, "message": "error in body parameter", "data": [ { "Message": "attributes parameter need an userID parameter", "Result": 0 } ] }`
		http.Error(w, respPar.Message, http.StatusBadRequest)
		return &responseParam{false, respPar.Message, "", "", "", "", nil, "", nil, ""}, err
	}
	respPar.UserID = mAttributes["userID"].(string)

	if mBody["type"] != nil {
		respPar.Type = mBody["type"].(string)
		if respPar.Type == "setbulk" {

			x := mBody["data"].(string)
			var xx []byte
			json.Unmarshal([]byte(x), &xx)

			fmt.Println(mBody["data"])
			fmt.Println("=========================================")
			fmt.Println(x)
			fmt.Println("=========================================")
			fmt.Println(xx)

			//respPar.Data = mBody["data"].([]byte)
		}
	}

	return &responseParam{true, "", respPar.App, respPar.Entity, respPar.Token, respPar.Function, respPar.Attributes, respPar.Type, respPar.Data, respPar.UserID}, err
}

//isJSON : Validate JSON Format
func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

//check : Check error
func check(err error, sInfo string) {
	if err != nil {
		fmt.Println("🐶"+sInfo, err.Error())
	}
}

//Key func for generate code
func Key() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err) // out of randomness, should never happen
	}
	return fmt.Sprintf("%x", buf)
	// or hex.EncodeToString(buf)
	// or base64.StdEncoding.EncodeToString(buf)
}

//Round : float64 ( x.long )
func Round(input float64, places int) (newVal float64) {
	var s string
	if places == 1 {
		s = fmt.Sprintf("%.1f", input)
	} else if places == 2 {
		s = fmt.Sprintf("%.2f", input)
	} else if places == 3 {
		s = fmt.Sprintf("%.3f", input)
	} else if places == 4 {
		s = fmt.Sprintf("%.4f", input)
	} else if places == 5 {
		s = fmt.Sprintf("%.5f", input)
	} else {
		s = fmt.Sprintf("%.6f", input)
	}
	iL := len(s)
	tTemp := []rune(s)
	x, err := strconv.Atoi(string(tTemp[iL-1 : iL]))
	if err != nil {
		x = 0
	}
	f64, err := strconv.ParseFloat(s, 64)
	if err != nil {
		f64 = input
	}

	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * f64
	if x == 0 {
		round = digit
	} else if x < 5 {
		round = math.Floor(digit)
	} else {
		round = math.Ceil(digit)
	}
	newVal = round / pow
	return
}

//RoundUp : float64 ( x.long )
func RoundUp(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
}

//RoundDown : float64 ( x.long )
func RoundDown(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Floor(digit)
	newVal = round / pow
	return
}
