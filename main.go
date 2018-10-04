package main

import (
	dss "github.com/c0ze/dss_reuters_go/api"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	//	log.SetLevel(log.WarnLevel)
	log.SetLevel(log.DebugLevel)
	dss.Init()

	// Convenience method for composite extractions
	location, status, _ := dss.OnDemandExtractComposite("KE1000001402")
	var result []byte
	for status == "InProgress" {
		time.Sleep(5 * time.Second)
		status, result = dss.GetAsyncResult(location)
	}
	log.Debug(string(result))

	// composite extraction with custom fields
	location, status, _ = dss.OnDemandExtract(
		"KE1000001402",
		"Composite",
		[]string{"Dividend Yield",
			"Domicile"},
		nil)

	for status == "InProgress" {
		time.Sleep(5 * time.Second)
		status, result = dss.GetAsyncResult(location)
	}
	log.Debug(string(result))

	// technical indicators extraction
	location, status, _ = dss.OnDemandExtract(
		"KE1000001402",
		"TechnicalIndicators",
		[]string{"Net Change - Close Price - 1 Day",
			"Percent Change - Close Price - 1 Day"},
		nil)

	for status == "InProgress" {
		time.Sleep(5 * time.Second)
		status, result = dss.GetAsyncResult(location)
	}
	log.Debug(string(result))

	// time series extraction
	location, status, _ = dss.OnDemandExtract(
		"KE1000001402",
		"TimeSeries",
		[]string{"Universal Close Price",
			"Trade Date"},
		map[string]string{
			"StartDate": "2018-01-01",
			"EndDate":   "2018-10-04"})

	for status == "InProgress" {
		time.Sleep(5 * time.Second)
		status, result = dss.GetAsyncResult(location)
	}
	log.Debug(string(result))

}
