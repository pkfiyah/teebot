package daemons

import (
	"context"
	"fmt"
	"time"

	"github.com/pkfiyah/tee1000/models"
	"github.com/pkfiyah/tee1000/teeonwrapper"
	"github.com/redis/go-redis/v9"
)

type TimeToFire struct {
	fireTime      *time.Time
	retryAttempts uint
	retry         bool
}

var ctx = context.Background()
var redClient = redis.NewClient(&redis.Options{
	Addr:     "teebot-redis-1:6379",
	Password: "",
	DB:       0,
})

func StartSnipingDaemon() {
	// redClient.Set(context.Background(), "test", "test", 0)
	// Check tee times in redis every 5 minutes
	go func(rClient *redis.Client) {
		for {
			fmt.Println("[SDaemon] Checking tee times")
			checkForTeeTimes()
			fmt.Println("[SDaemon] Check Complete, Sleeping 5 minutes")
			time.Sleep(5 * time.Minute)
		}
	}(redClient)
}

func checkForTeeTimes() {
	var cursor uint64
	iter := redClient.Scan(ctx, cursor, "TeeTime:*", 0).Iterator()
	for iter.Next(ctx) {
		fmt.Println("keys", iter.Val())
		res := redClient.Get(ctx, iter.Val())
		fmt.Printf("Redis Get: %v\n", res)
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	return
}

func snipeTeeTime() {
	fmt.Println("Daemon Loading Ammunition for Snipe")
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
