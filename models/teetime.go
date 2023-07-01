package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type TeeTime struct {
	BookingMember   *PlayerInfo
	BookingDate     string      //FORMAT: 2006-12-28
	TimesToSnipe    []time.Time //FORMAT: 2006-12-28;23:59
	NumPlayers      uint
	NumCarts        uint
	NumHoles        uint
	Retries         uint
	RedKey          string
	LastAttemptTime time.Time // FORMAT: 2006-12-28;23:59
}

var ctx = context.Background()
var redClient = redis.NewClient(&redis.Options{
	Addr:     "teebot-redis-1:6379",
	Password: "",
	DB:       0,
})

func GetTeeTimeByBooking(tT *TeeTime) (*TeeTime, error) {
	checkExisting, err := redClient.Get(ctx, getRedisKey(tT)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	exisingTeeTime := TeeTime{}
	err = json.Unmarshal([]byte(checkExisting), &exisingTeeTime)
	if err != nil {
		return nil, err
	}

	return &exisingTeeTime, nil
}

func SetTeeTimeWithBooking(r *http.Request, tT *TeeTime) error {

	// TODO Expire immediately after tee time has passed
	expTime := time.Until(time.Now().Add(time.Hour * 24 * 7))
	jsonTeeTime, err := json.Marshal(tT)
	if err != nil {
		return fmt.Errorf("Could not marshal data")
	}
	err = redClient.Set(r.Context(), fmt.Sprintf("TeeTime:%s/%s", tT.BookingMember.User.Fullname, tT.BookingDate), jsonTeeTime, expTime).Err()
	if err != nil {
		fmt.Printf("Err occurred saving tee time to Redis: %v\n", err)
	}
	return nil
}

func getRedisKey(tT *TeeTime) string {
	return fmt.Sprintf("TeeTime:%s/%s", tT.BookingMember.User.Fullname, tT.BookingDate)
}
