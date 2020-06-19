package main

import (
	"expertisetest/adnetwork"
	"expertisetest/config"
	"expertisetest/handler"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	config.GetInstance()
	h := handler.GetInstance()

	m, err := h.Load()
	if err != nil {
		logrus.Fatal(err)
		os.Exit(2)
	}

	if err := h.Store(m, true); err != nil {

		logrus.Fatal(err)
		os.Exit(2)
	}

	an := &adnetwork.AdNetwork{}
	a := config.GetInstance().RedisClient.Get("CN")
	if err := a.Scan(an); err != nil {
		logrus.Fatal(err)
		os.Exit(3)
	}

}
