package main

import (
	"fmt"

	"github.com/pkfiyah/tee1000/api"
	"github.com/pkfiyah/tee1000/daemons"
	"github.com/pkfiyah/tee1000/models"
)

func main() {
	fmt.Println("Loading player info...")
	err := models.LoadUserDataFromJson()
	if err != nil {
		fmt.Println("Bad player info, check user.json for proper location and structure")
		panic(err)
	}

	fmt.Println("Starting daemon...")
	daemons.StartSnipingDaemon()

	fmt.Println("Starting api...")
	api.HandleRequests()
}
