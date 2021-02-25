package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// MyEvent defines the event we get
type MyEvent struct {
	Bucket string `json:"Bucket"`
	Key    string `json:"Key"`
}

var myEvent MyEvent

func main() {
	content, err := ioutil.ReadFile("test.json")
	if err != nil {
		fmt.Println("Could not open 'test.json'")
		return
	}

	text := string(content)

	err = json.Unmarshal([]byte(text), &myEvent)
	if err != nil {
		fmt.Println("Could not parse event from 'test.json'")
		return
	}

	// { "Bucket": "bucket name", "Key": "key name", "waitTimeout": 5 }

	output := "{ \"Bucket\": \"" + myEvent.Bucket + "\", \"Key\": \"" + myEvent.Key + "\", \"waitTimeout\": 5 }"

	fmt.Println("test.json should look like: ")
	fmt.Println(output)

	fmt.Println("Got bucket: '" + myEvent.Bucket + "' and key: '" + myEvent.Key + "'")
}
