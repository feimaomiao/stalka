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
	database.Database_init(sugar)
	client, err := client.NewPandaClient(sugar)
	if err != nil {
		sugar.Fatal(err)
	}
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
	err = client.GetMatches()
	if err != nil {
		sugar.Fatal(err)
	}
	sugar.Infof("Done with initial setup, made %d requests", client.GetRun())
	sugar.Sync()
	matchTicker := time.NewTicker(2 * time.Hour)
	setupTicker := time.NewTicker(24 * time.Hour)
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
		sugar.Sync()
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
		sugar.Sync()
	}()
	for {
		time.Sleep(time.Hour * 1)
	}
}

// func main() {
// 	ticker := time.NewTicker(time.Second * 2)
// 	ticker2 := time.NewTicker(time.Second * 4)
// 	defer ticker.Stop()
// 	defer ticker2.Stop()
// 	go func() {
// 		for range ticker2.C {
// 			fmt.Println("Tick2")
// 		}
// 	}()
// 	go func() {
// 		for range ticker.C {
// 			fmt.Println("Tick")
// 		}
// 	}()
// 	for {
// 		time.Sleep(time.Hour * 1)
// 	}

// }
