package main

import (
	"fmt"
	"time"

	"github.com/pkfiyah/tee1000/api"
	"github.com/pkfiyah/tee1000/teeonwrapper"
)

func main() {
	toClient, err := teeonwrapper.NewTeeOnClient()
	if err != nil {
		fmt.Printf("Error creating Tee On client")
	}

	err = toClient.TeeOnSignIn()
	if err != nil {
		fmt.Printf("Err signing in to TeeOn via client")
	}

	err = toClient.TeeOnSnipeTime(time.Date(2023, 06, 21, 4, 0, 0, 0, time.Local))
	if err != nil {
		fmt.Printf("Err requesting tee time to TeeOn via client: %s", err)
	}

	api.HandleRequests()
}
