package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"bytes"
	"io"
	"os"
	//"encoding/json"
)

const (
	CONSUMER = ""
	SECRET = ""
)

func main() {
	getAuth(CONSUMER, SECRET)
	
}

func getAuth(login string, pass string) {
	//prepare consumer request
	//we should URL encode (RFC 1738)
	key := login + ":" + pass
	creds := base64.StdEncoding.EncodeToString([]byte(key))
	
	//get OAUTH token
	client := &http.Client{}
	
	endpoint := "https://api.twitter.com/oauth2/token"
	body := bytes.NewBufferString("grant_type=client_credentials")
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Basic " + creds)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	fmt.Println("Status: " + res.Status)
	
	//raw read for debug
	io.Copy(os.Stdout, res.Body)
	fmt.Println()
	
	/*
	var output struct{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&output)
	if err != nil {
		fmt.Println("JSON Decoding error:")
		//panic(err)
	}
	fmt.Println(output)
	*/
}
