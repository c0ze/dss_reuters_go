package data_scope

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

// Init provides setup for interacting with Reuters/Refinitiv
// DATA_SCOPE_USERNAME and DATA_SCOPE_PASSWORD credentials must be provided in ENV
// Saves access token for future requests.
//
// TODO: Handle Token expiration
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

type RequestType int

const (
	COMPOSITE RequestType = iota
	TECHNICAL_INDICATORS
	TIMESERIES
	INTRADAY_PRICING
)

type RequestStatus int

const (
	INIT RequestStatus = iota
	IN_PROGRESS
	COMPLETED
	FAILED
)

type IdentifierType int

const (
	RIC IdentifierType = iota
	ISIN
)

func (er *ExtractRequest) identifier() string {
	var idType string
	switch er.IdType {
	case ISIN:
		idType = "Isin"
	case RIC:
		idType = "Ric"
	default:
		idType = "Isin"
	}
	return fmt.Sprintf("{\"Identifier\": \"%s\",\"IdentifierType\": \"%s\"}", er.Identifier, idType)
}

func (er *ExtractRequest) setStatus(status string) {
	switch status {
	case "InProgress":
		er.Status = IN_PROGRESS
	case "":
		er.Status = COMPLETED
	default:
		er.Status = FAILED
	}
}

type ExtractRequest struct {
	ReqType RequestType
	Fields      []string
	IdType IdentifierType
	Identifier string
	Condition   map[string]string
	Status RequestStatus
	Location string
	Result []byte
}

func (er *ExtractRequest) TypeString() string {
	switch er.ReqType {
	case COMPOSITE:
		return "Composite"
	case TECHNICAL_INDICATORS:
		return "TechnicalIndicators"
	case TIMESERIES:
		return "TimeSeries"
	case INTRADAY_PRICING:
		return "IntradayPricing"
	default:
		return ""
	}
}

func (er *ExtractRequest) toString() string {
	fields, _ := json.Marshal(er.Fields)
	condition, _ := json.Marshal(er.Condition)
	return fmt.Sprintf(templ, er.TypeString(), fields, er.identifier(), condition)
}

// Extract sends an extract:on request to Reuters(Refinitiv)
// Returns location, status and body. Location is the URL which must
// be polled to retrieve the etraction result.
//
func (er *ExtractRequest) Extract() {
	extractURL := "/RestApi/v1/Extractions/ExtractWithNotes"

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

	er.setStatus(status)
	er.Location = location
}

// CheckResult tries to retrieve the result of an extraction request
// Result is saved in the ExtractRequest
//
func (er *ExtractRequest) CheckResult() {
	req, err := http.NewRequest("GET", er.Location, bytes.NewBuffer([]byte("")))
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
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body:", string(body))

	er.setStatus(status)
	if er.Status == COMPLETED {
		er.Result = body
	}
}
