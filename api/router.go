package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkfiyah/tee1000/models"
	"github.com/redis/go-redis/v9"
)

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Health Status: OK")
	fmt.Println("[Endpoint Hit]: health")
}

func addTeeTime(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/addTime" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	fmt.Println("Adding Tee Time")

	// Parse Values
	tT := r.FormValue("teeTime")
	date := strings.Split(tT, ";")[0]
	parsedTeeTime, err := time.Parse("2006-01-02;15:04", tT)
	if err != nil {
		fmt.Printf("Error parsing tee time from request: %s\n", err)
		return
	}

	carts := r.FormValue("carts")
	parsedCarts, err := strconv.ParseUint(carts, 10, 32)
	if err != nil {
		fmt.Printf("Error parsing carts from request: %s\n", err)
		return
	}

	players := r.FormValue("players")
	parsedPlayers, err := strconv.ParseUint(players, 10, 32)
	if err != nil {
		fmt.Printf("Error parsing players from request: %s\n", err)
		return
	}

	holes := r.FormValue("holes")
	parsedHoles, err := strconv.ParseUint(holes, 10, 32)
	if err != nil {
		fmt.Printf("Error parsing holes from request: %s\n", err)
		return
	}

	val := models.TeeTime{
		BookingMember: "Tylerfancy",
		Date:          date,
		TimesToSnipe:  []time.Time{parsedTeeTime},
		NumCarts:      uint(parsedCarts),
		NumPlayers:    uint(parsedPlayers),
		NumHoles:      uint(parsedHoles),
	}

	err = saveTeeTimeToRedis(r, &val)
	if err != nil {
		fmt.Printf("Error saving to redis")
	}

	fmt.Fprintf(w, "Tee Time Accepted")

}

func HandleRequests() {
	http.HandleFunc("/health", health)
	http.HandleFunc("/addTime", addTeeTime)
	log.Fatal(http.ListenAndServe(":9001", nil))
}

func saveTeeTimeToRedis(r *http.Request, teeTime *models.TeeTime) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "teebot-redis-1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// TODO Expire immediately after tee time has passed
	expTime := time.Until(time.Now().Add(time.Hour * 24 * 7))
	jsonTeeTime, err := json.Marshal(teeTime)
	if err != nil {
		return fmt.Errorf("Could not marshal data")
	}
	err = rdb.Set(r.Context(), fmt.Sprintf("TeeTime:%s/%s", teeTime.BookingMember, teeTime.Date), jsonTeeTime, expTime).Err()
	if err != nil {
		fmt.Printf("Err occurred saving tee time to Redis: %v\n", err)
	}
	return nil
}
