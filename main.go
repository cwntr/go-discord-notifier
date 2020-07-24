package main

import (
	"encoding/json"
	"fmt"
	"github.com/cwntr/go-discord-notifier/pkg/4chan"
	"github.com/cwntr/go-discord-notifier/pkg/common"
	"github.com/cwntr/go-discord-notifier/pkg/discord"
	"os"
	"time"
)

var (
	cfg common.Config
)

func main() {
	//parsing config -> local "cfg.json"
	err := readConfig()
	if err != nil {
		fmt.Printf("error reading config, err %v", err)
		os.Exit(1)
		return
	}

	//init notification channels
	discord.InitDiscordSession(cfg)

	//periodic workers
	registerJobs()

	select {} // run forever
}

func registerJobs() {
	go fourchan.PeriodicCheck(cfg)
}

// Reads the entire config, "cfg.json" is hardcoded and must be placed on same level as the application binary
func readConfig() error {
	file, err := os.Open("cfg.json")
	if err != nil {
		fmt.Printf("can't open config file: %v", err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Printf("can't decode config JSON: %v", err)
		return err
	}

	cfg.Interval, err = time.ParseDuration(cfg.ProcessInterval)
	if err != nil {
		return err
	}

	return nil
}
