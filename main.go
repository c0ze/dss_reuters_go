package main

import dss "github.com/c0ze/dss_reuters_go/api"

func main() {
	dss.Init()
	dss.OnDemandExtract("KE1000001402")
}
