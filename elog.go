package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	. "github.com/marstid/go-pdom"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// eLog Event Type
type Event struct {
	ID          string            `json:"uuid,omitempty" bson:"uuid"`
	Triggered   time.Time         `json:"triggered,omitempty" bson:"triggered"`
	Cleared     time.Time         `json:"cleared,omitempty" bson:"cleared,omitempty"`
	Fingerprint string            `json:"fingerprint" bson:"fingerprint"`
	Priority    string            `json:"priority,omitempty" bson:"priority,omitempty"` // P1, P2, P3, P4, P5
	Severity    string            `json:"severity,omitempty" bson:"severity,omitempty"` // Critical, Warning, Info
	Status      string            `json:"status,omitempty" bson:"status"`
	Msg         string            `json:"msg" bson:"msg"`
	Resource    string            `json:"resource,omitempty" bson:"resource,omitempty"`
	Source      string            `json:"source" bson:"source"`
	Site        string            `json:"site,omitempty" bson:"site,omitempty"`
	Env         string            `json:"env,omitempty" bson:"env,omitempty"` // Environment - Production, Stage, Test etc
	KB          string            `json:"kb,omitempty" bson:"kb,omitempty"`
	Ticket      string            `json:"ticket,omitempty" bson:"ticket,omitempty"`
	Comment     string            `json:"comment,omitempty" bson:"comment,omitempty"` // Operator comment added when acknowledge
	Ack         bool              `json:"ack,omitempty" bson:"ack,omitempty"`         // Used to mark event as acknowledge by operator
	AckBy       string            `json:"ackby,omitempty" bson:"ackby,omitempty"`
	GraphUrl    string            `json:"graphurl,omitempty" bson:"graphurl,omitempty"` // Optional: A Link to a graph associated with the event
	Tags        map[string]string `json:"tags,omitempty" bson:"tags,omitempty"`
}

func postToElog(url string, key string, check Check) error {

	var state string
	switch check.Status {
	case "down":
		state = "active"
	case "up":
		state = "resolved"
	default:
		state = "unknown"
	}

	x := make(map[string]string)

	err := json.Unmarshal([]byte(check.CustomMessage), &x)
	if err != nil {
		log.Printf("Unable to parse metadata from CustomMessage: %s\n", check.CustomMessage)
		log.Println("Alert will not be sent to eLog")
		return err
	}

	prio, ok := x["prio"]
	if !ok {
		prio = "P3"
	}

	msg, ok := x["msg"]
	if !ok {
		msg = ""
	}

	site, ok := x["site"]
	if !ok {
		site = ""
	}

	env, ok := x["env"]
	if !ok {
		env = ""
	}

	kb, ok := x["kb"]
	if !ok {
		kb = ""
	}

	graph, ok := x["graph"]
	if !ok {
		graph = fmt.Sprintf("https://my.pingdom.com/app/reports/uptime#check=%d", check.ID)
	}

	event := Event{
		Fingerprint: fmt.Sprintf("%d-%d", check.ID, check.Lasterrortime),
		//Fingerprint: fmt.Sprintf("%d", check.ID),
		Priority: prio,
		Site:     site,
		Tags:     x,
		Msg:      cfg.prepend + msg,
		Status:   state,
		Resource: "PINGDOM: " + check.Name,
		Source:   "PingdomPoller",
		Env:      env,
		KB:       kb,
		GraphUrl: graph,
	}

	postdata, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if cfg.debug {
		log.Println("POST Data:")
		log.Println(string(postdata))
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(postdata))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	var tr *http.Transport

	if cfg.insecureSSL {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{},
		}
	}

	client := &http.Client{Timeout: 5 * time.Second, Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.Status != "200 OK" {
		return fmt.Errorf("Post error:  %v: %v", resp.Status, string(body))
	}

	return nil
}
