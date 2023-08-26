package p2p

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTablePlayers(t *testing.T) {
	var (
		maxSeats = 10
		table    = NewTable(maxSeats)
	)

	for i := 0; i < maxSeats; i++ {
		assert.Nil(t, table.AddPlayer(strconv.Itoa(i)))
	}

	assert.Equal(t, maxSeats, table.LenPlayers())

	table.RemovePlayerByAddr("1")
	assert.Equal(t, maxSeats-1, table.LenPlayers())
}

func TestTablePlayerBefore(t *testing.T) {
	var (
		maxSeats = 10
		table    = NewTable(maxSeats)
	)
	assert.Nil(t, table.AddPlayer("1"))
	prevPlayer, err := table.GetPlayerBefore("1")
	assert.Nil(t, prevPlayer)
	assert.NotNil(t, err)

	assert.Nil(t, table.AddPlayer("2"))
	prevPlayer, err = table.GetPlayerBefore("2")
	assert.Nil(t, err)
	assert.Equal(t, "1", prevPlayer.addr)

	assert.Nil(t, table.AddPlayer("3"))
	assert.Nil(t, table.RemovePlayerByAddr("2"))
	prevPlayer, err = table.GetPlayerBefore("3")
	assert.Nil(t, err)
	assert.Equal(t, "1", prevPlayer.addr)
}

func TestTableGetPlayerAfter(t *testing.T) {
	var (
		maxSeats = 10
		table    = NewTable(maxSeats)
	)
	assert.Nil(t, table.AddPlayer("1"))
	assert.Nil(t, table.AddPlayer("2"))
	nextPlayer, err := table.GetPlayerAfter("1")
	assert.Nil(t, err)
	assert.Equal(t, "2", nextPlayer.addr)

	assert.Nil(t, table.AddPlayer("3"))
	assert.Nil(t, table.RemovePlayerByAddr("2"))
	nextPlayer, err = table.GetPlayerAfter("1")
	assert.Nil(t, err)
	assert.Equal(t, "3", nextPlayer.addr)

	nextPlayer, err = table.GetPlayerAfter("3")
	assert.Nil(t, err)
	assert.Equal(t, "1", nextPlayer.addr)

	assert.Nil(t, table.RemovePlayerByAddr("1"))
	nextPlayer, err = table.GetPlayerAfter("3")
	assert.Nil(t, nextPlayer)
	assert.NotNil(t, err)
}

func TestTableRemovePlayer(t *testing.T) {
	var (
		maxSeats = 10
		table    = NewTable(maxSeats)
	)

	for i := 0; i < maxSeats; i++ {
		addr := fmt.Sprintf(":%d", i)
		assert.Nil(t, table.AddPlayer(addr))
		assert.Nil(t, table.RemovePlayerByAddr(addr))

		player, err := table.GetPlayer(addr)
		assert.NotNil(t, err)
		assert.Nil(t, player)
	}

	assert.Equal(t, 0, table.LenPlayers())
}

func TestTableAddPlayer(t *testing.T) {
	var (
		maxSeats = 1
		table    = NewTable(maxSeats)
	)

	assert.Nil(t, table.AddPlayer(":1"))
	assert.Equal(t, 1, table.LenPlayers())

	assert.NotNil(t, table.AddPlayer(":2"))
	assert.Equal(t, 1, table.LenPlayers())
}

func TestTableGetPlayer(t *testing.T) {
	var (
		maxSeats = 10
		table    = NewTable(maxSeats)
	)

	for i := 0; i < maxSeats; i++ {
		addr := fmt.Sprintf(":%d", i)
		assert.Nil(t, table.AddPlayer(addr))
		player, err := table.GetPlayer(addr)
		assert.Nil(t, err)
		assert.Equal(t, player.addr, addr)
	}

	assert.Equal(t, maxSeats, table.LenPlayers())
}
