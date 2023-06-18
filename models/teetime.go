package models

import "time"

type TeeTime struct {
	BookingMember string
	TimesToSnipe  []time.Time
	NumPlayers    uint
	NumCarts      uint
	NumHoles      uint
}
