/*
* @Author: wangqilong
* @Description:
* @File: http
* @Date: 2021/9/15 11:44 下午
 */

package request

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func Get(url string) (body []byte, httpCode int, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return body, httpCode, fmt.Errorf("http newRequest: %v", err.Error())
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return body, httpCode, fmt.Errorf("http client do: %v", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, resp.StatusCode, fmt.Errorf("read response body: %v", err.Error())
	}
	return body, resp.StatusCode, nil
}

func Post(url string, data []byte) (body []byte, httpCode int, err error) {
	request, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return body, httpCode, fmt.Errorf("HTTP POST NewRequest: %v URL: %v", err.Error(), url)
	}
	defer request.Body.Close()

	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return body, httpCode, fmt.Errorf("HTTP POST client.Do: %v, url: %s body: %v", err.Error(), url, request)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("HTTP POST: %v, url: %v, body: %v", url, err.Error(), string(respBytes))
	}

	return respBytes, resp.StatusCode, nil
}
