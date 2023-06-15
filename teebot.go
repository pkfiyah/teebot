package main

import (
	"fmt"
	"pkfiyah/tee1000/teeonwrapper"
	"time"
)

func main() {
	client, err := teeonwrapper.NewTeeOnClient()
	if err != nil {
		fmt.Printf("Error creating Tee On client")
	}

	err = client.TeeOnSignIn()
	if err != nil {
		fmt.Printf("Err signing in to TeeOn via client")
	}

	err = client.TeeOnSnipeTime(time.Date(2023, 06, 21, 8, 0, 0, 0, time.Local))
	if err != nil {
		fmt.Printf("Err requesting tee time to TeeOn via client: %s", err)
	}
}
