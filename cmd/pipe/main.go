package main

import (
	"encoding/json"
	"expertisetest/adnetwork"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	network := adnetwork.GenerateList()

	out, _ := json.MarshalIndent(struct {
		Data []*adnetwork.AdNetwork `json:"data"`
	}{
		Data: network,
	}, "", "  ")
	fmt.Println(string(out))
}
