package db

import (
	"fmt"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"msgqueue-luke.com/v2/internals/utils"
)

var store *StoreMain = NewStoreMain()

// var randStrx string = utils.RandomString(5)
var randStr2 string = utils.RandomString(5)
var randStr3 string = utils.RandomString(5)

func TestCreate(t *testing.T) {
	var randStr1 string = utils.RandomString(5)
	err := store.Create(randStr1, randStr1)
	require.NoError(t, err)

	// val before update
	beforeNode := store.Fetch(randStr1)
	nodeVal := beforeNode.Value
	dateBefore := beforeNode.UpdatedAt

	require.Equal(t, randStr1, nodeVal)

	err = store.Update(randStr1, "cussoke", time.Now().Add(5*time.Minute))
	require.NoError(t, err)

	node := store.Fetch(randStr1)
	nodeValAfter := node.Value
	dateAfter := node.UpdatedAt
	// val after update
	require.Equal(t, "cussoke", node.Value)

	// val after and before must be different
	require.False(t, nodeVal == nodeValAfter, "must be false")

	// updateAt after must be younger
	require.Greater(t, dateAfter, dateBefore, "check date")

	log.Printf("before: %s and after: %s", nodeVal, nodeValAfter)

}

func TestCreateNegative(t *testing.T) {

	_ = store.Create(randStr3, randStr3)
	err := store.Create(randStr3, randStr3)
	require.Error(t, err, "err must be appeared, bcs data exists")

	err = store.Update("notfoundidentifier", "notfound", time.Now())
	log.Printf("err update not found: %s", err.Error())
	require.Error(t, err, fmt.Sprintf("err must be appeared, because not found: %s", err.Error()))

	err = store.Update(randStr2, "anyval", time.Now())
	require.Error(t, err, "err expected: gagal update compare swap, karena nilai old sudah diganti goroutine lain")

	nodeVal := store.Fetch("notfounid")
	require.Nil(t, nodeVal, "must be nil")

}

func TestUpdate_CompareAndSwapFail(t *testing.T) {
	s := &StoreMain{}
	key := "key-race"

	s.mutexMap.Store(key, &Node{
		Key:       key,
		Value:     "initial",
		UpdatedAt: time.Now(),
	})

	const n = 10
	start := make(chan struct{})
	done := make(chan error, n)

	successCount := int32(0)
	failCount := int32(0)

	for i := 0; i < n; i++ {
		go func(i int) {
			<-start // semua goroutine nunggu barrier, biar start bareng
			err := s.Update(key, fmt.Sprintf("value-%d", i), time.Now())
			done <- err
		}(i)
	}

	close(start) // lepas semua goroutine sekaligus

	for i := 0; i < n; i++ {
		// semua di sini return nil, tujuan test cuma untuk eksekusi baris log.Printf
		select {
		case msg1 := <-done:
			if msg1 != nil {
				atomic.AddInt32(&failCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}

		}
	}

	require.Greater(t, successCount, int32(0), "success count harus lebih dari 0")
	log.Printf("success: %d", successCount)
	log.Printf("fail: %d", failCount)

}
