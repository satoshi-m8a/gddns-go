package main

import (
	"encoding/json"
	"os"
	"fmt"
	"net/http"
	"io/ioutil"
	"bytes"
	"strings"
)

type Conf struct {
	URL         string
	TOKEN       string
	SECRET      string
	ZONE        string
	DOMAIN_NAME string
	TTL         int
}

type Credential struct {
	token  string
	secret string
}

type Zone struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	CurrentVersionId string `json:"current_version_id"`
}

type Zones []Zone

type RecordValue struct {
	Address string `json:"address"`
}

type RecordValues []RecordValue

type Record struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	TTL         int `json:"ttl"`
	Name        string `json:"name"`
	EnableAlias bool `json:"enable_alias"`
	Records     RecordValues `json:"records"`
}

type Records []Record

func loadConf(path string) Conf {
	file, _ := os.Open(path)
	decoder := json.NewDecoder(file)
	conf := Conf{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("error:", err)
	}
	return conf
}

func getZone(url, targetZone string, cred Credential) Zone {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cred.token, cred.secret)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
	}
	zones := make(Zones, 0)
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &zones)
	defer res.Body.Close()
	var zone Zone
	for _, v := range zones {
		if v.Name == targetZone {
			zone = v
		}
	}
	return zone
}

func getRecord(url, targetDomain string, zone Zone, cred Credential) Record {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url + "/" + zone.Id + "/versions/" + zone.CurrentVersionId + "/records", nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cred.token, cred.secret)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
	}
	records := make(Records, 0)
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &records)
	defer res.Body.Close()
	var record Record
	for _, v := range records {
		if v.Name == targetDomain {
			record = v
		}
	}
	return record
}

func updateRecord(url string, zone Zone, record Record, cred Credential) Record {
	println(record.Records[0].Address)
	r, _ := json.Marshal(record)
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", url + "/" + zone.Id + "/versions/" + zone.CurrentVersionId + "/records/" + record.Id, bytes.NewBuffer(r))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cred.token, cred.secret)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
	}
	var newRec Record
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &newRec)
	defer res.Body.Close()
	return newRec
}

func getGlobalIp() string {
	res, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		fmt.Println("error:", err)
	}
	ip, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	return strings.TrimSpace(string(ip))
}

func main() {
	if len(os.Args) < 2 {
		println("usege: gddns PATH_TO_CONFJSON")
		println("example: gddns ./conf.json")
		return
	}
	path := os.Args[1]
	conf := loadConf(path)
	cred := Credential{token:conf.TOKEN, secret:conf.SECRET}
	zone := getZone(conf.URL, conf.ZONE, cred)
	record := getRecord(conf.URL, conf.DOMAIN_NAME, zone, cred)
	ip := getGlobalIp()

	if record.Records[0].Address == ip {
		println("ip is:" + ip)
	} else {
		record.Records[0].Address = ip
		record.TTL = conf.TTL
		newRecord := updateRecord(conf.URL, zone, record, cred)
		println("new ip is:" + newRecord.Records[0].Address)
	}

}
