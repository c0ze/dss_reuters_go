package main

import (
	dss "github.com/c0ze/dss_reuters_go/api"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	dss.Init()
	dss.OnDemandExtract("KE1000001402")
}
