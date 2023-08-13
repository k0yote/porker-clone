package deck

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type Suit int

func (s Suit) String() string {
	switch s {
	case Spades:
		return "SPADES"
	case Harts:
		return "HARTES"
	case Diamonds:
		return "DIAMONDS"
	case Clubs:
		return "CLUBS"
	default:
		panic("invalid Suit")
	}
}

const (
	Spades Suit = iota
	Harts
	Diamonds
	Clubs
)

type Card struct {
	Suit  Suit
	Value int
}

func (c Card) String() string {
	value := strconv.Itoa(c.Value)
	if c.Value == 1 {
		value = "ACE"
	}
	return fmt.Sprintf("%s of %s %s", value, c.Suit, suitToUnicode(c.Suit))
}

func NewCard(s Suit, v int) Card {
	if v > 13 {
		panic("the value of the card cannot be greater than 13")
	}

	return Card{
		Suit:  s,
		Value: v,
	}
}

type Deck [52]Card

func New() Deck {
	var (
		nSuits = 4
		nCrads = 13
		d      = [52]Card{}
	)

	x := 0
	for i := 0; i < nSuits; i++ {
		for j := 0; j < nCrads; j++ {
			d[x] = NewCard(Suit(i), j+1)
			x++
		}
	}

	return shuffle(d)
}

func shuffle(d Deck) Deck {

	rand.NewSource(time.Now().UnixNano())

	rand.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})

	return d
}

func suitToUnicode(s Suit) string {
	switch s {
	case Spades:
		return "♠"
	case Harts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	default:
		panic("invalid Suit")
	}
}
