package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
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

	fmt.Fprintf(w, "Tee Time Accepted")
	fmt.Println("Adding Tee Time")

	// Parse Values
	tT := r.FormValue("teeTime")
	parsedTeeTime, err := time.Parse("2006-01-02;15:04", tT)
	if err != nil {
		fmt.Printf("Error parsing tee time from request: %s\n", err)
	}

	carts := r.FormValue("carts")
	parsedCarts, err := strconv.ParseUint(carts, 10, 32)
	if err != nil {
		fmt.Printf("Error parsing carts from request: %s\n", err)
	}

	players := r.FormValue("players")
	parsedPlayers, err := strconv.ParseUint(players, 10, 32)
	if err != nil {
		fmt.Printf("Error parsing players from request: %s\n", err)
	}

	holes := r.FormValue("holes")
	parsedHoles, err := strconv.ParseUint(holes, 10, 32)
	if err != nil {
		fmt.Printf("Error parsing holes from request: %s\n", err)
	}

	val := models.TeeTime{
		BookingMember: "Tylerfancy",
		TimesToSnipe:  []time.Time{parsedTeeTime},
		NumCarts:      uint(parsedCarts),
		NumPlayers:    uint(parsedPlayers),
		NumHoles:      uint(parsedHoles),
	}

	err = saveTeeTimeToRedis(r, &val)
	if err != nil {
		fmt.Printf("Error saving to redis")
	}

}

func HandleRequests() {
	http.HandleFunc("/health", health)
	http.HandleFunc("/addTime", addTeeTime)
	log.Fatal(http.ListenAndServe(":9001", nil))
}

func saveTeeTimeToRedis(r *http.Request, teeTime *models.TeeTime) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	expTime := time.Until(time.Now().Add(time.Hour * 24 * 7))
	err := rdb.Set(r.Context(), teeTime.BookingMember, teeTime, expTime).Err()
	if err != nil {
		fmt.Printf("Err occurred saving tee time to Redis: %v\n", err)
	}
	return nil
}
