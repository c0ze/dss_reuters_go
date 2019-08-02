package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	dataScopeUsername string
	dataScopePassword string
	baseURI           string
	loginResp         LoginResponse
)

type LoginResponse struct {
	Context string `json:"@odata.context"`
	Token   string `json:"value"`
}

func Init() {
	dataScopeUsername := os.Getenv("DATA_SCOPE_USERNAME")
	dataScopePassword := os.Getenv("DATA_SCOPE_PASSWORD")
	baseURI = "https://hosted.datascopeapi.reuters.com"

	log.Debug("credentials: ", dataScopeUsername, dataScopePassword)

	loginUrl := "/RestApi/v1/Authentication/RequestToken"
	log.Debug("URL:>", loginUrl)

	credentials := map[string]map[string]string{"Credentials": {"Username": dataScopeUsername, "Password": dataScopePassword}}
	jsonCredentials, _ := json.Marshal(credentials)
	log.Debug("BODY:>", string(jsonCredentials))

	req, err := http.NewRequest("POST", baseURI+loginUrl, bytes.NewBuffer(jsonCredentials))
	req.Header.Set("Prefer", "respond-async")
	req.Header.Set("Content-Type", "application/json; odata=minimalmetadata")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Debug("login response: status ", resp.Status)
	log.Debug("login response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		log.Debug("whoops:", err)
	}
	log.Debug("Token:", loginResp.Token)
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
    "Condition": %s
  }
}`

type ExtractRequest struct {
	RequestType string
	Fields      []string
	Identifiers map[string]string
	Condition   map[string]string
}

func (er *ExtractRequest) toString() string {
	fields, _ := json.Marshal(er.Fields)
	identifiers, _ := json.Marshal(er.Identifiers)
	condition, _ := json.Marshal(er.Condition)
	return fmt.Sprintf(templ, er.RequestType, fields, identifiers, condition)
}

func OnDemandExtractComposite(isinCode string) (string, string, []byte) {
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

	log.Debug("request Body:", er.toString())
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

	log.Debug("response Status:", resp.Status)
	log.Debug("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body:", string(body))

	location := resp.Header.Get("Location")
	status := resp.Header.Get("Status")

	log.Debug("location: ", location)
	log.Debug("status: ", status)

	return location, status, body
}

func OnDemandExtract(isinCode string, requestType string, fields []string, condition map[string]string) (string, string, []byte) {
	extractURL := "/RestApi/v1/Extractions/ExtractWithNotes"

	er := ExtractRequest{
		RequestType: requestType,
		Fields:      fields,
		Identifiers: map[string]string{
			"Identifier":     isinCode,
			"IdentifierType": "Isin"},
		Condition: condition}

	log.Debug("request Body:", er.toString())
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

	log.Debug("response Status:", resp.Status)
	log.Debug("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body:", string(body))

	location := resp.Header.Get("Location")
	status := resp.Header.Get("Status")

	log.Debug("location: ", location)
	log.Debug("status: ", status)

	return location, status, body
}

func GetAsyncResult(location string) (string, []byte) {
	req, err := http.NewRequest("GET", location, bytes.NewBuffer([]byte("")))
	req.Header.Set("Prefer", "respond-async; wait=5")
	req.Header.Set("Content-Type", "application/json; odata=minimalmetadata")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", loginResp.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Debug("response Status:", resp.Status)
	log.Debug("response Headers:", resp.Header)
	status := resp.Header.Get("Status")
	log.Debug("status: ", status)
	if status == "InProgress" {
		return status, []byte("")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Debug("response Body:", string(body))
		return status, body
	}
}