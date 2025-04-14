package main

import (
	"time"

	"github.com/feimaomiao/stalka/client"
	"github.com/feimaomiao/stalka/database"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	err := database.Init(sugar)
	if err != nil {
		sugar.Fatal(err)
	}
	client, err := client.NewPandaClient(sugar)
	if err != nil {
		sugar.Fatal(err)
	}
	err = client.Startup()
	if err != nil {
		sugar.Fatal(err)
	}
	day := 24
	matchTicker := time.NewTicker(time.Hour)
	setupTicker := time.NewTicker(time.Duration(day) * time.Hour)
	defer matchTicker.Stop()
	defer setupTicker.Stop()
	go func() {
		for range matchTicker.C {
			sugar.Info("Matchticker fired")
			err = client.GetMatches()
			if err != nil {
				sugar.Fatal(err)
			}
			sugar.Infof("Done with run, made %d requests so far", client.GetRun())
		}
	}()
	go func() {
		for range setupTicker.C {
			sugar.Info("Setupticker fired")
			err = client.UpdateGames()
			if err != nil {
				sugar.Fatal(err)
			}
			err = client.GetLeagues()
			if err != nil {
				sugar.Fatal(err)
			}
			err = client.GetSeries()
			if err != nil {
				sugar.Fatal(err)
			}
			err = client.GetTournaments()
			if err != nil {
				sugar.Fatal(err)
			}
			sugar.Infof("Done with setup, made %d requests so far", client.GetRun())
		}
	}()
	for {
		time.Sleep(time.Hour)
	}
}
