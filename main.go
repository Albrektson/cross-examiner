package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"bytes"
	"io"
	"os"
	"encoding/json"
)

const (
	CONSUMER = ""
	SECRET = ""
	USER = ""
)

type msg struct {
	Text string
	ID int
}

func main() {
	access_token := getAuth(CONSUMER, SECRET)
	msgList := getMessages(access_token, USER)
	
	parseMessages(msgList)
}

func parseMessages(msgList []msg) {
	for _, m := range msgList {
	}
}

func debugBody (input io.Reader) {
	io.Copy(os.Stdout, input)
}

//fetches the last 3200 tweets from a given user
func getMessages(token string, user string) []msg {
	//build & send twitter userinfo request
	client := &http.Client{}
	endpoint := "https://api.twitter.com/1.1/statuses/user_timeline.json"
	endpoint += "?screen_name=" + user
	endpoint += "&include_rts=false"
	req, err := http.NewRequest("GET", endpoint, nil)
	
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer " + token)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	
	if res.StatusCode == 200 {
		//decode json response
		var messages []msg
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&messages)
		if err != nil {
			panic(err)
		}
		
		return messages
	} else {
		fmt.Println(res)
		panic(res.Status)
	}
}

func getAuth(login string, pass string) string {
	//prepare consumer request
	//we should URL encode (RFC 1738)
	key := login + ":" + pass
	creds := base64.StdEncoding.EncodeToString([]byte(key))
	
	//build & send OAUTH token request
	client := &http.Client{}
	endpoint := "https://api.twitter.com/oauth2/token"
	body := bytes.NewBufferString("grant_type=client_credentials")
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Basic " + creds)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	
	//decode json response
	var output struct{
		Errors string
		Token_type string
		Access_token string
	}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&output)
	if err != nil {
		fmt.Println("Status: " + res.Status)
		fmt.Println("JSON Decoding error:")
		panic(err)
	}
	if res.StatusCode != 200 {
		panic(output.Errors)
	}
	
	return output.Access_token
}