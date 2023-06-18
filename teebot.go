package main

import (
	"fmt"
	"time"

	"github.com/pkfiyah/tee1000/daemons"
	"github.com/pkfiyah/tee1000/models"
	"github.com/pkfiyah/tee1000/teeonwrapper"
)

func main() {
	//api.HandleRequests()
	//startSniperDaemon()
	snipeTeeTime()
}

func startSniperDaemon() {
	go func() {
		for {
			daemons.CheckForTeeTimes()
			time.Sleep(24 * time.Hour)
		}
	}()
}

func snipeTeeTime() {
	toClient, err := teeonwrapper.NewTeeOnClient()
	if err != nil {
		fmt.Printf("Error creating Tee On client")
	}

	err = toClient.TeeOnSignIn()
	if err != nil {
		fmt.Printf("Err signing in to TeeOn via client")
	}

	times := []time.Time{
		time.Date(2023, 06, 24, 7, 30, 0, 0, time.Local),
		time.Date(2023, 06, 24, 7, 39, 0, 0, time.Local),
		time.Date(2023, 06, 24, 7, 48, 0, 0, time.Local),
		time.Date(2023, 06, 24, 7, 57, 0, 0, time.Local),
		time.Date(2023, 06, 24, 8, 06, 0, 0, time.Local),
		time.Date(2023, 06, 24, 8, 15, 0, 0, time.Local),
		time.Date(2023, 06, 24, 8, 24, 0, 0, time.Local),
	}

	tT := &models.TeeTime{
		TimesToSnipe: times,
		NumPlayers:   4,
		NumCarts:     0,
	}

	err = toClient.TeeOnSnipeTime(tT)
	if err != nil {
		fmt.Printf("Err requesting tee time to TeeOn via client: %s", err)
	}
}
