package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func LoadNodeTable() map[string]string {
	a, err := ioutil.ReadFile("nodetable.csv")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	NodeTable := map[string]string{}
	err = json.Unmarshal(a, &NodeTable)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if len(NodeTable) == N {
		return NodeTable
	} else {
		fmt.Println("[Error]: inconsistent nodetable length with N!")
		return nil
	}
}
