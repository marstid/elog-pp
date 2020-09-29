package main

import (
	"fmt"
	. "github.com/marstid/go-pdom"
	"log"
	"os"
	"testing"
	"time"
)

func TestElog(t *testing.T) {
	fmt.Println("Test: Client Connect")
	checkEnv()

	client, err := NewRestClient(os.Getenv("PINGDOM_TOKEN"), false, 45)
	if err != nil {
		t.Error(err)
	}

	check, err := client.UptimeGetCheckDetails(4811721)
	if err != nil {
		t.Error(err)
	}

	check.Status = "down"
	check.Lasterrortime = int(time.Now().Unix())
	err = postToElog(os.Getenv("WEBHOOK_URL"), os.Getenv("WEBHOOK_TOKEN"), check)
	if err != nil {
		t.Error(err)
	}

	check.Status = "up"
	err = postToElog(os.Getenv("WEBHOOK_URL"), os.Getenv("WEBHOOK_TOKEN"), check)
	if err != nil {
		t.Error(err)
	}

	log.Println(check.CustomMessage)

}
