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
			fmt.Println("[SnipeD] Checking tee times")
			checkForTeeTimes()
			fmt.Println("[SnipeD] Check Complete, Sleeping 5 minutes")
			time.Sleep(5 * time.Minute)
		}
	}(redClient)
}

func checkForTeeTimes() {
	var cursor uint64
	iter := redClient.Scan(ctx, cursor, "TeeTime:*", 0).Iterator()
	for iter.Next(ctx) {
		fmt.Printf("[SnipeD] Tee Time Found: %v\n", iter.Val())
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
		fmt.Printf("NextTime: %s", parsedTeeTime.NextAttemptTime)
		fmt.Printf("NowTime: %s", time.Now().Local())
		fmt.Printf("CheckTime: %s", time.Now().Local().Sub(parsedTeeTime.NextAttemptTime))
		if time.Now().Local().Sub(parsedTeeTime.NextAttemptTime) > 0 {
			// Load found teeTime in for firing
			loadMagazine(parsedTeeTime)
		}
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	return
}

func loadMagazine(ammo *models.TeeTime) {
	fmt.Println("[SnipeD]Loading Ammunition for Snipe")
	toClient, err := teeonwrapper.NewTeeOnClient()
	if err != nil {
		fmt.Printf("Error creating Tee On client")
	}

	err = toClient.TeeOnSignIn(ammo.BookingMember)
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

			// This indicates booking will unlock shortly
			if waitTime != nil {
				fmt.Printf("Err: %s", err)
				if err == teeonwrapper.ErrBookingNotAvailable {
					// Might not be able to get this booking, update retries in redis
					ammo.Retries++
					ammo.LastAttemptTime = time.Now().Local()
					byteData, err := json.Marshal(ammo)
					if err != nil {
						fmt.Errorf("Uh Oh") // TODO better error handle here
						return
					}
					redClient.Set(ctx, ammo.RedKey, byteData, time.Until(time.Now().Add(time.Hour*24*7)))
				}
				fmt.Printf("CheckInn WaitTime:%s\n", *waitTime)
				fmt.Printf("CheckInn Duyraitions:%s\n", time.Duration(time.Minute*5))
				fmt.Printf("Resujhtl:%v\n", *waitTime < time.Duration(time.Minute*5))
				if err != teeonwrapper.ErrBookingNotAvailable && *waitTime < time.Duration(time.Minute*5) {
					fmt.Printf("WERWERWR")
					time.Sleep(*waitTime)
				}
				if err != teeonwrapper.ErrBookingNotAvailable && *waitTime >= time.Duration(time.Minute*5) {
					fmt.Printf("FGFGHFGJHGF")
					loc, _ := time.LoadLocation("America/Halifax")
					ammo.LastAttemptTime = time.Now().In(loc)
					fmt.Printf("Wait this long: %v\n", *waitTime-time.Duration(time.Second*30))
					ammo.NextAttemptTime = time.Now().In(loc).Add(*waitTime - time.Duration(time.Second*30))
					ammo.Retries = 0
					byteData, err := json.Marshal(ammo)
					if err != nil {
						fmt.Errorf("Uh Oh") // TODO better error handle here
						return
					}
					redClient.Set(ctx, ammo.RedKey, byteData, time.Until(time.Now().Add(time.Hour*24*7)))
				}
			}

			// Not close enough to booking to want to try again currently
			if err != nil && waitTime == nil && err == teeonwrapper.ErrTooEarlyToRegisterTeeTime {
				fmt.Println("Need to wait a bit before reattempting, booking not open yet")
				magazineEmptied = true
				return
			}

			if err != nil && err == teeonwrapper.ErrTeeTimeAlreadyBooked {
				fmt.Println("Booking already complete, no need to continue")
				magazineEmptied = true
				redClient.Del(ctx, ammo.RedKey)
				return
			}

			// TODO currently can get here if we have an error thats not captured, will purge these entries. Clean this part up
			if err == nil {
				magazineEmptied = true
				redClient.Del(ctx, ammo.RedKey)
				return
			}

			ammo.Retries++
			fmt.Printf("Attempting Retry: %d\n", ammo.Retries)
		}
	}()
}
