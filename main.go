package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	authEmail        = "your-cloudflare-email"
	authKey          = "your-auth-key"
	updateZoneName   = "the-cloudflare-zone-name(e.g. example.com)"
	updateRecordName = "the-record-name(e.g. subdomain.example.com)"
	baseAPIURL       = "https://api.cloudflare.com/client/v4"
	zonesAPI         = "%s/zones"
	recordsAPI       = "%s/zones/%s/dns_records"
	recordsUpdateAPI = "%s/zones/%s/dns_records/%s"
)

var client *http.Client

// Cloudflare API Response Structs
type zones struct {
	Result []zone
}

type zone struct {
	ID     string
	Name   string
	Status string
}

func (z zones) Get(i int) zone {
	return z.Result[i]
}

type records struct {
	Result []record
}

type record struct {
	ID         string `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	Content    string `json:"content,omitempty"`
	Proxiable  bool   `json:"proxiable,omitempty"`
	Proxied    bool   `json:"proxied,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	Locked     bool   `json:"locked,omitempty"`
	ZoneID     string `json:"zone_id,omitempty"`
	ZoneName   string `json:"zone_name,omitempty"`
	CreatedOn  string `json:"created_on,omitempty"`
	ModifiedOn string `json:"modified_on,omitempty"`
}

type recordUpdate struct {
	Success bool
}

func (r records) Get(i int) record {
	return r.Result[i]
}

func main() {
	client = &http.Client{}

	z, err := getZone(updateZoneName)
	if err != nil {
		fmt.Println("[Error] Problem retrieving the zone")
		log.Fatal(err)
		return
	}

	rec, err := getRecord(updateRecordName, z.ID)
	if err != nil {
		fmt.Println("[Error] Problem retrieving the record")
		log.Fatal(err)
		return
	}

	ip, err := getExternalIP()
	if err != nil {
		fmt.Println("[Error] Problem retrieving the external ip")
		log.Fatal(err)
		return
	}

	success, err := updateRecordIP(ip, rec)
	if err != nil {
		fmt.Println("[Error] Problem updating the record ip")
		log.Fatal(err)
		return
	}

	fmt.Println(success)
}

// getAPIJSON prepares and sends a GET request to Cloudflare API
func getAPIJSON(url string, target interface{}) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Auth-Email", authEmail)
	req.Header.Add("X-Auth-Key", authKey)

	res, err := client.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}

// putAPIJSON prepares and sends a PUT request to Cloudflare API
func putAPIJSON(data []byte, url string, target interface{}) error {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	req.Header.Add("X-Auth-Email", authEmail)
	req.Header.Add("X-Auth-Key", authKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}

// getZone retrieves Zone data from the Cloudflare API
func getZone(s string) (zone, error) {
	results := &zones{}
	apiURL := fmt.Sprintf(zonesAPI, baseAPIURL)
	err := getAPIJSON(apiURL, results)
	if err != nil {
		return zone{}, err
	}

	for i := range results.Result {
		zone := results.Get(i)
		if zone.Name == s {
			return zone, nil
		}
	}

	return zone{}, fmt.Errorf("requested zone name, '%s', was not found", s)
}

// getZone retrieves Record data from the Cloudflare API
func getRecord(s, zoneID string) (record, error) {
	results := &records{}
	apiURL := fmt.Sprintf(recordsAPI, baseAPIURL, zoneID)
	err := getAPIJSON(apiURL, results)
	if err != nil {
		return record{}, err
	}

	for i := range results.Result {
		record := results.Get(i)
		if record.Name == s {
			return record, nil
		}
	}

	return record{}, fmt.Errorf("requested dns record name '%s' was not found for the zone id '%s'", s, zoneID)
}

// updateRecordIP updates the IP for a Zone Record via the Cloudflare API
func updateRecordIP(pointTo string, rec record) (string, error) {
	if pointTo == rec.Content {
		return "IP has not changed, no need to update DNS", nil
	}

	rec.Content = pointTo
	data, err := json.Marshal(rec)
	if err != nil {
		return "", err
	}

	result := &recordUpdate{}
	apiURL := fmt.Sprintf(recordsUpdateAPI, baseAPIURL, rec.ZoneID, rec.ID)
	err = putAPIJSON(data, apiURL, result)
	if err != nil {
		return "", err
	}

	if result.Success {
		return fmt.Sprintf("DNS IP Successfully Updated to '%v' for record '%v'", pointTo, rec.Name), nil
	}

	return "", fmt.Errorf("Unsuccessful updating DNS to '%v' for record '%v'", pointTo, rec.Name)
}

// getExternalIP retrieves the public ip address for your server
func getExternalIP() (string, error) {
	res, err := http.Get("http://checkip.amazonaws.com")

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contents)), nil
}