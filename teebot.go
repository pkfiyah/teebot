package main

import (
	"fmt"
	"time"

	"github.com/pkfiyah/tee1000/api"
	"github.com/pkfiyah/tee1000/daemons"
	"github.com/pkfiyah/tee1000/models"
	"github.com/pkfiyah/tee1000/teeonwrapper"
)

func main() {
	fmt.Println("Starting daemon")
	daemons.StartSnipingDaemon()

	// snipeTeeTime()
	fmt.Println("Starting api")
	api.HandleRequests()
}

// func startSniperDaemon() {
// 	// Create redis connection for daemon
// 	redClient := redis.NewClient(&redis.Options{
// 		Addr:     "teebot-redis-1:6379",
// 		Password: "",
// 		DB:       0,
// 	})
// 	//redClient.Set(context.Background(), "test", "test", 0)
// 	// Check tee times in redis every 5 minutes
// 	go func(rClient *redis.Client) {
// 		for {
// 			fmt.Println("[Daemon] Checking tee times")
// 			daemons.CheckForTeeTimes(rClient)
// 			fmt.Println("[Daemon] Check Complete, Sleeping 5 minutes")
// 			time.Sleep(5 * time.Minute)
// 		}
// 	}(redClient)
//}

func snipeTeeTime() {
	fmt.Println("Sniping time")
	toClient, err := teeonwrapper.NewTeeOnClient()
	if err != nil {
		fmt.Printf("Error creating Tee On client")
	}

	err = toClient.TeeOnSignIn()
	if err != nil {
		fmt.Printf("Err signing in to TeeOn via client")
	}

	times := []time.Time{
		time.Date(2023, 06, 27, 6, 54, 0, 0, time.Local),
		// time.Date(2023, 06, 24, 7, 39, 0, 0, time.Local),
		// time.Date(2023, 06, 24, 7, 48, 0, 0, time.Local),
		// time.Date(2023, 06, 24, 7, 57, 0, 0, time.Local),
		// time.Date(2023, 06, 24, 8, 06, 0, 0, time.Local),
		// time.Date(2023, 06, 24, 8, 15, 0, 0, time.Local),
		// time.Date(2023, 06, 24, 8, 24, 0, 0, time.Local),
	}

	tT := &models.TeeTime{
		BookingMember: "Tylerfancy",
		TimesToSnipe:  times,
		NumPlayers:    4,
		NumCarts:      0,
		NumHoles:      18,
	}

	err = toClient.TeeOnSnipeTime(tT)
	if err != nil {
		fmt.Printf("Err requesting tee time to TeeOn via client: %s", err)
	}
	fmt.Println("Sniping time finished")
}
