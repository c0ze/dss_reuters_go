package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	dssUsername string
	dssPassword string
	baseURI     string
	loginResp   LoginResponse
)

type LoginResponse struct {
	Context string `json:"@odata.context"`
	Token   string `json:"value"`
}

func Init() {
	dssUsername := os.Getenv("DSS_USERNAME")
	dssPassword := os.Getenv("DSS_PASSWORD")
	baseURI = "https://hosted.datascopeapi.reuters.com"

	fmt.Println(dssUsername, dssPassword)

	loginUrl := "/RestApi/v1/Authentication/RequestToken"
	fmt.Println("URL:>", loginUrl)

	credentials := map[string]map[string]string{"Credentials": {"Username": dssUsername, "Password": dssPassword}} // }"username": username, "password": password}
	jsonCredentials, _ := json.Marshal(credentials)
	//	loginBody := []byte(fmt.Sprintf(`{"Credentials": {"Username": "%s","Password": "%s"}}`, dssUsername, dssPassword))
	fmt.Println("BODY:>", string(jsonCredentials))

	req, err := http.NewRequest("POST", baseURI+loginUrl, bytes.NewBuffer(jsonCredentials))
	req.Header.Set("Prefer", "respond-async")
	req.Header.Set("Content-Type", "application/json; odata=minimalmetadata")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	fmt.Println("Token:", loginResp.Token)
}

const templ = `{
  "ExtractionRequest": {
    "@odata.type": "#ThomsonReuters.Dss.Api.Extractions.ExtractionRequests.%sExtractionRequest",
    "ContentFieldNames": %s,
    "IdentifierList": {
      "@odata.type": "#ThomsonReuters.Dss.Api.Extractions.ExtractionRequests.InstrumentIdentifierList",
      "InstrumentIdentifiers": [%s],
      "ValidationOptions": null,
      "UseUserPreferencesForValidationOptions": false
    },
    "Condition": null
  }
}`

type ExtractRequest struct {
	RequestType string
	Fields      []string
	Identifiers map[string]string
}

func (er *ExtractRequest) toString() string {
	fields, _ := json.Marshal(er.Fields)
	identifiers, _ := json.Marshal(er.Identifiers)
	return fmt.Sprintf(templ, er.RequestType, fields, identifiers)
}

func OnDemandExtract(isinCode string) {
	// def initialize(session, fields, identifiers, type)

	extractURL := "/RestApi/v1/Extractions/ExtractWithNotes"

	er := ExtractRequest{
		RequestType: "Composite",
		Fields: []string{"Close Price",
			"Contributor Code Description",
			"Currency Code Description",
			"Dividend Yield",
			"Main Index",
			"Market Capitalization",
			"Market Capitalization - Local Currency",
			"Percent Change - Close Price - 1 Day",
			"Universal Close Price Date"},
		Identifiers: map[string]string{
			"Identifier":     isinCode,
			"IdentifierType": "Isin"}}

	fmt.Println("request Body:", er.toString())
	req, err := http.NewRequest("POST", baseURI+extractURL, bytes.NewBuffer([]byte(er.toString())))
	req.Header.Set("Prefer", "respond-async; wait=5")
	req.Header.Set("Content-Type", "application/json; odata=minimalmetadata")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", loginResp.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("response Body:", string(body))

}
