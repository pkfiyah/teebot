package daemons

import (
	"context"
	"encoding/json"
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

// Check tee times in redis every 5 minutes
func StartSnipingDaemon() {
	go func(rClient *redis.Client) {
		for {
			fmt.Println("[SDaemon] Checking tee times")
			checkForTeeTimes()
			fmt.Println("[SDaemon] Check Complete, Sleeping 5 minutes")
			time.Sleep(5 * time.Minute)
		}
	}(redClient)
}

func loadMagazine(ammo *models.TeeTime) {
	fmt.Println("Daemon Loading Ammunition for Snipe")
	toClient, err := teeonwrapper.NewTeeOnClient()
	if err != nil {
		fmt.Printf("Error creating Tee On client")
	}

	err = toClient.TeeOnSignIn()
	if err != nil {
		fmt.Printf("Err signing in to TeeOn via client")
	}

	// Once we have a client thats signed in, can start making tee time requests
	// Here we will take shots and based on the results of those, three results can happen:
	// 1. The teeTime is too far off -> Unlock is >= 5 minutes in the future, so we'll let the daemon retry it later
	// 2. The teeTime is < 5 minutes from an unlock, or other potential conflicts that are retryable have occured
	// 			Typically here we will let the inspector return a duration to retry the times at, which will be < the 5 minute retry time
	// 3. We have successfully booked the tee time, no retry needed
	go func() {
		magazineEmptied := false
		for !magazineEmptied && ammo.Retries < 10 {
			waitTime, err := toClient.TeeOnSnipeTime(ammo)
			if err != nil && waitTime != nil {
				if err == teeonwrapper.ErrTooEarlyToRegisterTeeTime {
					fmt.Println("Too Early to book, sleeping for a bit")
					time.Sleep(*waitTime)
				}
			}

			if err != nil && err == teeonwrapper.ErrTooEarlyToRegisterTeeTime {
				fmt.Println("Need to wait a bit before reattemtpign, booking not open uyet")
				magazineEmptied = true
				// redClient.Del(ctx, ammo.RedKey)
				return
			}

			if err != nil && err == teeonwrapper.ErrTeeTimeAlreadyBooked {
				fmt.Println("Booking already complete, no need to continue")
				magazineEmptied = true
				redClient.Del(ctx, ammo.RedKey)
				return
			}

			if err == nil {
				magazineEmptied = true
				redClient.Del(ctx, ammo.RedKey)
				return
			}

			ammo.Retries++
			fmt.Printf("Attempting Retry: %d\n", ammo.Retries)
		}
		ammo.Retries = 0
	}()
}

func checkForTeeTimes() {
	var cursor uint64
	iter := redClient.Scan(ctx, cursor, "TeeTime:*", 0).Iterator()
	for iter.Next(ctx) {
		fmt.Printf("Tee time Found: %v\n", iter.Val())
		res := redClient.Get(ctx, iter.Val())

		// Check sniping times for found results
		resBytes, err := res.Bytes()
		if err != nil {
			fmt.Printf("Err with turning redis result into bytes")
			return
		}

		parsedTeeTime := &models.TeeTime{}
		if err = json.Unmarshal(resBytes, parsedTeeTime); err != nil {
			fmt.Printf("Error unmarshalling tee times from redis")
			return
		}
		parsedTeeTime.RedKey = iter.Val()
		// Load found teeTime in for firing
		loadMagazine(parsedTeeTime)
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	return
}
