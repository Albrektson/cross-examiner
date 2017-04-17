package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode"
)

const (
	CONSUMER = ""
	SECRET = ""
	USER1     = "cnn"
	USER2     = "bbc"
	ALLOW_RETWEETS = "false"
	TEST     = FINGERPRINT
)

const (
	MESSAGE = iota
	WORD
	FINGERPRINT
	ANGULAR
)

type msg struct {
	Text           string
	ID             int
	NormalizedText string
	Tokens         []string
}

func main() {
	access_token := getAuth(CONSUMER, SECRET)
	msgList1 := getMessages(access_token, USER1)
	msgList2 := getMessages(access_token, USER2)
	parseMessages(msgList1)
	parseMessages(msgList2)

	dummyMsg1 := msg{Text: "Adding more messages to timeline.", ID: -1}
	dummyMsg2 := msg{Text: "Adding more messages to timeline now.", ID: -2}
	dummyMsg3 := msg{Text: "Adding some messages to timeline.", ID: -3}
	dummyMsg4 := msg{Text: "Selecting messages to include in timeline.", ID: -4}
	dummyList := []msg{dummyMsg1, dummyMsg2, dummyMsg3, dummyMsg4}
	parseMessages(dummyList)

	switch TEST {
	case MESSAGE:
		dumbCompare(msgList1, dummyList)
	case WORD:
		wordCompare(msgList1, dummyList)
	case FINGERPRINT:
		//rand.Seed(time.Now)
		fingerprintCompare(msgList1, msgList2)
	case ANGULAR:
		fmt.Println("Angular comparison method under construction")
	default:
		panic("No test case chosen")
	}
}

func angularCompare() {

}

//selects fingerprints from messages in one list
//and looks for them in messages from the second list
func fingerprintCompare(msgList1 []msg, msgList2 []msg) {
	for _, m1 := range msgList1 {
		t1 := m1.NormalizedText
		t1Len := len(t1)
		fpSize := int(t1Len / 4)
		anchor := rand.Intn(t1Len-fpSize)
		fingerprint := t1[anchor:(anchor+fpSize)]
		for _, m2 := range msgList2 {
			t2 := m2.NormalizedText
			if len(t2) <= fpSize {
				fmt.Println("Found message too short for fingerprinting, ignoring.")
				continue
			}
			if strings.Contains(t2, fingerprint) {
				fmt.Println("Found messages with fingerprinting similarity.")
				fmt.Printf("Fingerprint was [%s].\n", fingerprint)
				fmt.Printf("Message 1: [%s]\tID: [%d]\n", m1.Text, m1.ID)
				fmt.Printf("Message 2: [%s]\tID: [%d]\n\n", m2.Text, m2.ID)
			}
		}
	}
}

//compares messages word by word for similarity, which means
//any offset can make near identical messages entirely dissimilar
func wordCompare(msgList1 []msg, msgList2 []msg) {
	for _, m1 := range msgList1 {
		for _, m2 := range msgList2 {
			wordCount := 0
			duplicates := 0
			len1 := len(m1.Tokens)
			len2 := len(m2.Tokens)
			if len1 < len2 {
				wordCount = len2
				for i, w1 := range m1.Tokens {
					w2 := m2.Tokens[i]
					if w1 == w2 {
						duplicates++
					}
				}
			} else {
				wordCount = len1
				for i, w1 := range m2.Tokens {
					w2 := m1.Tokens[i]
					if w1 == w2 {
						duplicates++
					}
				}
			}
			//fmt.Printf("Duplicates: %d, Wordcount: %d\n", duplicates, wordCount)
			if float64(duplicates)/float64(wordCount) > 0.5 {
				fmt.Println("Found messages with high similarity rating:")
				fmt.Printf("Message 1: [%s]\tID: [%d]\n", m1.Text, m1.ID)
				fmt.Printf("Message 2: [%s]\tID: [%d]\n", m2.Text, m2.ID)
			}
		}
	}
}

//compares if any message in msgList equals another message
func dumbCompare(msgList1 []msg, msgList2 []msg) {
	for _, m1 := range msgList1 {
		for _, m2 := range msgList2 {
			if m1.Text == m2.Text {
				fmt.Println("Found duplicate messages:")
				fmt.Printf("Text: [%s]\tID: [%d]\n", m1.Text, m1.ID)
				fmt.Printf("Text: [%s]\tID: [%d]\n", m2.Text, m2.ID)
			}
		}
	}
}

//goes over a list of tweets normalizing and tokenizing text
func parseMessages(msgList []msg) {
	for i, m := range msgList {
		text := m.Text

		hashtag, _ := regexp.Compile("#")
		//hashtag, _ := regexp.Compile("#[A-z0-9]+")
		text = hashtag.ReplaceAllString(text, "")

		usertag, _ := regexp.Compile("(\\.)*@[A-z0-9]+")
		text = usertag.ReplaceAllString(text, "")

		//we should maybe replace with "<link>" instead?
		webaddr, _ := regexp.Compile("([A-z0-9]+\\.)*([A-z0-9]+\\.[A-z]{2,})(/[A-z0-9]*)*")
		text = webaddr.ReplaceAllString(text, "")

		whitespace, _ := regexp.Compile("[[:space:]]")
		text = whitespace.ReplaceAllString(text, " ")
		
		m.NormalizedText = text

		//FieldsFunc: string -> []string, using the given delimiter
		m.Tokens = strings.FieldsFunc(m.NormalizedText, isSpecialChar)
		msgList[i] = m
	}
}

//returns true if char c is not a letter of the alphabet
func isSpecialChar(c rune) bool {
	return !unicode.IsLetter(c)
}

//debug function for HTTP data reading
func debugBody(input io.Reader) {
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
	req.Header.Add("Authorization", "Bearer "+token)
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
	req.Header.Add("Authorization", "Basic "+creds)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	//decode json response
	var output struct {
		Errors       string
		Token_type   string
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
