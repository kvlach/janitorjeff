package rps

import (
	"math/rand"
	"time"
)

const (
	paper = iota
	scissors
	rock
)

const (
	win = iota
	draw
	loss
)

func run(player int) (int, int) {
	var computer int
	rand.Seed(time.Now().UnixNano())
	switch rand.Intn(3) {
	case 0:
		computer = rock
	case 1:
		computer = paper
	case 2:
		computer = scissors
	}

	var result int
	if player == computer {
		result = draw
	} else if player == (computer+1)%3 {
		// the winning choice is always positioned to the right, for example
		// paper = 0, scissors beats papers, scissors = 1. Mod 3 for rock = 2
		// which beats paper.
		result = win
	} else {
		result = loss
	}

	return result, computer
}
