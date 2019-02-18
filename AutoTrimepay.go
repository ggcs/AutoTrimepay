package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var email = ""      //Trimepay账户
var password = ""   //密码
var method = "1"    //1:支付宝  2:微信
var alipayLimit = 0 //支付宝提现上限，单位分
var wechatLimit = 0 //微信提现上限，单位分

func main() {
	urlHome := "https://api.trimepay.com/"
	client := &http.Client{}
	rand.Seed(time.Now().UnixNano())
	addLog(time.Now().String(), false)

	letterRunes := []rune(`abcdefghijklmnopqrstuvwxyz`)
	csrf := make([]rune, 32)
	for csrfIndex := range csrf {
		csrf[csrfIndex] = letterRunes[rand.Intn(len(letterRunes))]
	}
	requestBody := url.Values{}
	requestBody.Set("email", email)
	requestBody.Set("password", password)
	request, _ := http.NewRequest(
		"POST",
		urlHome+"passport/auth/login?CSRF="+string(csrf),
		strings.NewReader(requestBody.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, _ := client.Do(request)
	cookies := response.Cookies()
	responseBody, _ := ioutil.ReadAll(response.Body)
	var responseBodyMap map[string]interface{}
	errorLog := json.Unmarshal(responseBody, &responseBodyMap)
	if errorLog != nil {
		addLog(errorLog.Error(), true)
	}
	if responseBodyMap["code"].(float64) != 0 {
		addLog("Login fail", true)
	}
	errorLog = response.Body.Close()
	if errorLog != nil {
		addLog(errorLog.Error(), true)
	}

	request, _ = http.NewRequest(
		"GET",
		urlHome+"merchant/app/dashboard?CSRF="+string(csrf),
		nil)
	for _, cookieIndex := range cookies {
		request.AddCookie(cookieIndex)
	}
	response, _ = client.Do(request)
	responseBody, _ = ioutil.ReadAll(response.Body)

	errorLog = json.Unmarshal(responseBody, &responseBodyMap)
	if errorLog != nil {
		addLog(errorLog.Error(), true)
	}

	balance := int(responseBodyMap["data"].(map[string]interface{})["merchant"].(map[string]interface{})["balance"].(float64))
	errorLog = response.Body.Close()
	if errorLog != nil {
		addLog(errorLog.Error(), true)
	}

	if balance <= 0 {
		addLog("No enough Balance", true)
	}

	if method == "1" && balance > alipayLimit {
		balance = alipayLimit
	}
	if method == "2" && balance > wechatLimit {
		balance = wechatLimit
	}

	requestBody = url.Values{}
	requestBody.Set("withdrawMethod", method)
	requestBody.Set("totalFee", strconv.Itoa(balance))
	request, _ = http.NewRequest(
		"POST",
		urlHome+"merchant/withdraw/create?CSRF="+string(csrf),
		strings.NewReader(requestBody.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookieIndex := range cookies {
		request.AddCookie(cookieIndex)
	}
	response, errorLog = client.Do(request)
	if errorLog != nil {
		addLog(errorLog.Error(), true)
	}
	errorLog = response.Body.Close()
	if errorLog != nil {
		addLog(errorLog.Error(), true)
	}
}

var allLog = ""

func addLog(log string, exit bool) {
	allLog += log
	allLog += "\n"

	if exit {
		allLog += "\n"
		logFile, _ := os.OpenFile("AutoTrimepay.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			os.ModeAppend)
		logFile.WriteString(allLog)
		os.Exit(0)
	}
}
