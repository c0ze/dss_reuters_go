package main

import (
	"github.com/c0ze/dss_reuters_go/api/data_scope"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	//	log.SetLevel(log.WarnLevel)
	log.SetLevel(log.DebugLevel)

	data_scope.Init()

	er := data_scope.ExtractRequest{
		ReqType: data_scope.COMPOSITE,
		Fields:	[]string{"Dividend Yield",
			"Domicile"},
		IdType: data_scope.ISIN,
		Identifier: "US4592001014",
		Condition: nil}

	er.Extract()

	for er.Status == data_scope.IN_PROGRESS {
		time.Sleep(5 * time.Second)
		er.CheckResult()
	}
	log.Debug(er.Status)
	log.Debug(er.Result)

	// // technical indicators extraction
	// location, status, _ = data_scope.OnDemandExtract(
	// 	"KE1000001402",
	// 	"TechnicalIndicators",
	// 	[]string{"Net Change - Close Price - 1 Day",
	// 		"Percent Change - Close Price - 1 Day"},
	// 	nil)

	// for status == "InProgress" {
	// 	time.Sleep(5 * time.Second)
	// 	status, result = data_scope.GetAsyncResult(location)
	// }
	// log.Debug(string(result))

	// // time series extraction
	// location, status, _ = data_scope.OnDemandExtract(
	// 	"KE1000001402",
	// 	"TimeSeries",
	// 	[]string{"Universal Close Price",
	// 		"Trade Date"},
	// 	map[string]string{
	// 		"StartDate": "2018-01-01",
	// 		"EndDate":   "2018-10-04"})

	// for status == "InProgress" {
	// 	time.Sleep(5 * time.Second)
	// 	status, result = data_scope.GetAsyncResult(location)
	// }
	// log.Debug(string(result))

}
