package main

import (
	"TwoRoundBB/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	nodeTable := config.LoadNodeTable()
	jsonMsg, err := json.Marshal("begin")
	if err != nil {
		fmt.Println(err)
	}

	for _, u := range nodeTable {
		url := u + "/begin"
		buff := bytes.NewBuffer(jsonMsg)
		fmt.Printf("Send to: %v, with msg: %v\n", url, buff)
		_, err1 := http.Post("http://"+url, "application/json", buff)
		if err1 != nil {
			fmt.Println(err1)
			//return
		}
	}

	fmt.Println("Consensus Started!")
}
