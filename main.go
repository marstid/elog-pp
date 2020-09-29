package main

import (
	"fmt"
	. "github.com/marstid/go-pdom"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	version string
	build   string
)

var pingdomChecks []Check

type CheckData struct {
	Status      string
	FingerPrint string
}

type Config struct {
	debug       bool
	url         string
	token       string
	prepend     string
	downgrade   bool
	insecureSSL bool
}

var cfg Config

func main() {
	log.Printf("Starting eLog Pingdom Poller version %s built %s", version, build)

	history := make(map[int]Check, 25)

	checkEnv()

	// Set up a channel to listen to for interrupt signals
	var runChan = make(chan os.Signal, 1)

	// Handle ctrl+c/ctrl+x interrupt
	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	scrapeInterval, err := time.ParseDuration(os.Getenv("PINGDOM_INT"))
	if err != nil {
		if cfg.debug {
			log.Print(os.Getenv("PINGDOM_INT"))
			log.Println(err)
		}

		// Set default interval
		scrapeInterval, _ = time.ParseDuration("60s")
	}

	ticker := time.NewTicker(scrapeInterval)
	done := make(chan bool)

	// Get Client
	client, err := NewRestClient(os.Getenv("PINGDOM_TOKEN"), false, 45)
	if err != nil {
		log.Println(err.Error())
		ticker.Stop()
		done <- true
	}

	// Test Client
	_, err = client.UptimeGetChecks()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("APA")
		ticker.Stop()
		done <- true
		os.Exit(1)

	}

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				pingdomChecks, err := client.UptimeGetDownChecksMap()
				if err != nil {
					fmt.Println(err.Error())
					return
					//os.Exit(1)
				}

				// Clear history
				for id, check := range history {
					_, ok := pingdomChecks[id]
					if ok {
						// Alert is still active - No action
					} else {
						// Alert in not active - Clear Alert
						//check.Status = "up"
						det, err := client.UptimeGetCheckDetails(id)
						if err != nil {
							log.Println(err)
						}

						check.Status = "up"
						check.CustomMessage = det.CustomMessage
						err = postToElog(os.Getenv("WEBHOOK_URL"), os.Getenv("WEBHOOK_TOKEN"), check)
						if err != nil {
							log.Println(err)
						}
						log.Printf("Clear: %d: %s %s \n", check.ID, check.Name, check.CustomMessage)
					}
				}

				// Trigger New Alerts
				for id, check := range pingdomChecks {
					_, ok := history[id]
					if ok {
						// Alert already in history - No action
					} else {
						// Alert not in history - Trigger Alert

						// Get the Custom message only avilable in detail call to pingdom
						detailed, err := client.UptimeGetCheckDetails(id)
						if err != nil {
							log.Println(err)
						} else {
							check.CustomMessage = detailed.CustomMessage
						}

						err = postToElog(os.Getenv("WEBHOOK_URL"), os.Getenv("WEBHOOK_TOKEN"), check)
						if err != nil {
							log.Println(err)
						}
						log.Printf("Trigger: %d: %s - %s \n", check.ID, check.Name, check.CustomMessage)
					}
				}

				history = pingdomChecks

				if cfg.debug {

					for i, check := range history {
						log.Printf("Debug: %d: %s - %s \n", i, check.Name, check.Status)

					}
					log.Println("Tick at", t)
				}
			}
		}
	}()

	// Block on this channel listeninf for those previously defined syscalls assign
	// to variable so we can let the user know why the server is shutting down
	interrupt := <-runChan
	log.Printf("Server is shutting down due to %+v\n", interrupt)

	ticker.Stop()
	done <- true

}

func checkEnv() {
	if os.Getenv("PINGDOM_TOKEN") == "" {
		fmt.Println("PINGDOM_TOKEN not set")
		os.Exit(1)
	}

	// Optional - Default to 60s
	if os.Getenv("PINGDOM_INT") == "" {
		fmt.Println("PINGDOM_INT not set - Using default 60s")
		//os.Exit(1)
	}

	if os.Getenv("WEBHOOK_TOKEN") == "" {
		fmt.Println("WEBHOOK_TOKEN not set")
		os.Exit(1)
	}
	cfg.url = os.Getenv("WEBHOOK_URL")

	if os.Getenv("WEBHOOK_URL") == "" {
		fmt.Println("WEBHOOK_URL not set")
		os.Exit(1)
	}
	cfg.token = os.Getenv("WEBHOOK_TOKEN")

	if os.Getenv("PP_DEBUG") == "1" {
		cfg.debug = true
	} else {
		cfg.debug = false
	}

	if os.Getenv("PP_INSECURE") == "1" {
		cfg.insecureSSL = true
	} else {
		cfg.insecureSSL = false
	}

	cfg.prepend = os.Getenv("PP_PREPEND")
}
