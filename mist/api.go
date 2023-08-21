package mist

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type PAPIConfigResponse struct {
	ID             string // device id
	SiteID         uuid.UUID
	OrgID          uuid.UUID
	MistConfigured bool     // whether we're configuring
	Config         string   // XML - unused
	ConfigCmd      []string // CLI commands
	Timestamp      float64  // config generated timestamp, used as version string
	Debugging      bool
	SecurityLog    SecurityLog `json:"SecurityLog"`
	IPSec          struct {
		CACerts     []string   // CA certs
		ClientCerts []struct { // Client Certs
			Name string
			Cert string
			Key  string
		}
	}
}

type SecurityLog struct {
	Enabled         bool     `json:"Enabled"`       // whether security log is enabled
	SourceAddress   string   `json:"SourceAddress"` // If configured, dynamic source address detection is overridden by user config
	Host            string   `json:"Host"`          // Destination IP address, srx-log-term.mist.com
	Port            uint16   `json:"Port"`          // Destination Port number to connect, 55514
	ServerCACerts   []string `json:"ServerCACerts"` // Server CAs, for mTLS
	ClientCACerts   []string `json:"ClientCACerts"` // Client CAs, for mTLS
	ClientCert      string   `json:"ClientCert"`    // Client/Device Certificate
	ClientKey       string   `json:"ClientKey"`     // Client Private key
	SourceInterface string   `json:"SourceInterface"`
}

func GetDeviceCerts(mac string) (SecurityLog, error) {
	url := fmt.Sprintf("http://papi-internal-staging.mist.pvt/internal/devices/%s/config", mac)
	log.Println(url)
	body, err := getHTTPResponse(url)
	if err != nil {
		return SecurityLog{}, err
	}
	var papiResponse PAPIConfigResponse
	err = json.Unmarshal(body, &papiResponse)
	if err != nil || len(papiResponse.SecurityLog.ServerCACerts) == 0 {
		return SecurityLog{}, err
	}
	//log.Println(papiResponse.SecurityLog)
	return papiResponse.SecurityLog, nil
}

func getHTTPResponse(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return b, err
}
