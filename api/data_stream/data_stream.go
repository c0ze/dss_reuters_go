package data_stream

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	dataStreamUsername string
	dataStreamPassword string
	baseURI            string
	loginResp          LoginResponse
)

type LoginResponse struct {
	TokenValue  string `json:"TokenValue"`
	TokenExpiry string `json:"TokenExpiry"`
}

// Init provides setup for interacting with Reuters/Refinitiv
// DATA_STREAM_USERNAME and DATA_STREAM_PASSWORD credentials must be provided in ENV
// Saves access token for future requests.
//
// TODO: Handle Token expiration
func Init() {
	dataStreamUsername := os.Getenv("DATA_STREAM_USERNAME")
	dataStreamPassword := os.Getenv("DATA_STREAM_PASSWORD")
	baseURI = "http://product.datastream.com"

	log.Debug("credentials: ", dataStreamUsername, dataStreamPassword)

	loginUrl := "/DSWSClient/V1/DSService.svc/rest/Token"
	log.Debug("URL:>", loginUrl)

	req, err := http.NewRequest("GET", baseURI+loginUrl, nil)
	req.Header.Set("Prefer", "respond-async")
	req.Header.Set("Content-Type", "application/json; odata=minimalmetadata")

	q := req.URL.Query()
	q.Add("UserName", dataStreamUsername)
	q.Add("Password", dataStreamPassword)
	req.URL.RawQuery = q.Encode()

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
	log.Debug("Token:", loginResp.TokenValue)
}

const templ = `{
  "DataRequest": {
    "Date": {
      "Start": "%s",
      "End": "%s",
      "Frequency": "",
      "Kind": 1
    },
    "Instrument": {
      "Properties": [
        {
          "Key": "IsSymbolSet",
          "Value": true
        }
      ],
      "Value": "%s"
    },
    "Tag": null
  },
  "Properties": null,
  "TokenValue": "%s"
}`

type IdentifierType int

const (
	RIC IdentifierType = iota
	ISIN
)

func (sr *StreamRequest) identifier() string {
	switch sr.Type {
	case ISIN:
		return sr.Identifier
	case RIC:
		return fmt.Sprintf("<%s>", sr.Identifier)
	default:
		return sr.Identifier
	}
}

// This is a special type to parse Refinitiv date formats
// eg. Date(1514764800000+0000) (should be 2018-01-01)
type RefinitivDate struct {
	time.Time
}

var rgx = regexp.MustCompile(`\((.*?)\+`)

func (sd *RefinitivDate) UnmarshalJSON(input []byte) error {
	s := string(input)
	strInput := rgx.FindStringSubmatch(s)

	// TODO properly handle zone information in the timestamp
	i, err := strconv.ParseInt(strInput[0][1:11], 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)

	sd.Time = tm
	return nil
}

type StreamRequest struct {
	StartDate  string
	EndDate    string
	Type       IdentifierType
	Identifier string
}

type StreamResponse struct {
	DataResponse struct {
		AdditionalResponses []struct {
			Key   string
			Value string
		}
		DataTypeNames  string
		DataTypeValues []struct {
			DataType     string
			SymbolValues []struct {
				Currency string
				Symbol   string
				Type     int
				Value    []decimal.Decimal
			}
		}
		Dates       []RefinitivDate
		SymbolNames string
		Tag         string
	}
	Properties string
}

func (sr *StreamRequest) toString() string {
	return fmt.Sprintf(templ, sr.StartDate, sr.EndDate, sr.identifier(), loginResp.TokenValue)
}

// Stream fetches the data from refinitiv.
// Result is saved in the Stream Request
//
// TODO: save the result in the original Request instead of creating a new one
func Stream(sr StreamRequest) StreamResponse {
	streamURL := "/DswsClient/V1/DSService.svc/rest/GetData"

	log.Debug("request Body:", sr.toString())
	req, err := http.NewRequest("POST", baseURI+streamURL, bytes.NewBuffer([]byte(sr.toString())))
	req.Header.Set("Content-Type", "application/json")

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

	var sResp StreamResponse
	err = json.Unmarshal(body, &sResp)
	if err != nil {
		log.Debug("whoops:", err)
	}
	log.Debug("Response:", sResp)
	return sResp
}
