package main

import (
	"fmt"
	"splunkTcp/s2s"
	utils "splunkTcp/util"
	"time"
)

func main() {
	client := s2s.S2S{}
	err := utils.ReadYAMLFile("s2s_config.yaml", &client)
	fmt.Println("Error: ", err)
	client.Connect()
	event := map[string]string{
		"source": "source",
		"text":   "text",
		"host":   "host",
	}
	client.AutoPush(5 * time.Second)
	for i := 0; i < 60; i++ {
		client.Add(event, "test")
		time.Sleep(1 * time.Second)
	}

}
