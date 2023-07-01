package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkfiyah/tee1000/models"
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
	parsedTeeTime, err := time.Parse("2006-01-02;15:04", tT)
	if err != nil {
		fmt.Printf("Error parsing tee time from request: %s\n", err)
		return
	}

	bookingDate := strings.Split(parsedTeeTime.Format("2006-01-02;15:04"), ";")
	if len(bookingDate) != 2 {
		fmt.Printf("Error parsing booking time from request")
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

	playerInfo, err := models.GetPlayerInfo()

	val := &models.TeeTime{
		BookingMember: playerInfo,
		BookingDate:   bookingDate[0],
		TimesToSnipe:  []time.Time{parsedTeeTime},
		NumCarts:      uint(parsedCarts),
		NumPlayers:    uint(parsedPlayers),
		NumHoles:      uint(parsedHoles),
		Retries:       uint(0),
	}

	existingBooking, err := models.GetTeeTimeByBooking(val)
	if err != nil {
		fmt.Printf("Error getting existing booking from redis: %s\n", err)
		return
	}

	if existingBooking != nil {
		// Add new time to existing booking instead of creating new booking
		existingBooking.TimesToSnipe = append(existingBooking.TimesToSnipe, val.TimesToSnipe...)
		val = existingBooking
	}

	err = models.SetTeeTimeWithBooking(r, val)
	if err != nil {
		fmt.Printf("Error saving to redis")
		return
	}

	fmt.Fprintf(w, "Tee Time Accepted")

}

func HandleRequests() {
	http.HandleFunc("/health", health)
	http.HandleFunc("/addTime", addTeeTime)
	log.Fatal(http.ListenAndServe(":9001", nil))
}
