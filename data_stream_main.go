package main

import (
	data_stream "github.com/c0ze/dss_reuters_go/api/data_stream"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	//	log.SetLevel(log.WarnLevel)
	log.SetLevel(log.DebugLevel)

	data_stream.Init()
	sr := data_stream.StreamRequest{
		Type:       data_stream.ISIN,
		Identifier: "US4592001014",
		StartDate:  "2018-01-01",
		EndDate:    "2018-01-04"}

	sResp := data_stream.Stream(sr)
	log.Debug(sResp)

	sr = data_stream.StreamRequest{
		Type:       data_stream.RIC,
		Identifier: "IBM.N",
		StartDate:  "2018-01-01",
		EndDate:    "2018-01-04"}

	sResp = data_stream.Stream(sr)
	log.Debug(sResp)

}
