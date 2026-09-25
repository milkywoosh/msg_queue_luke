package db

import (
	"fmt"
	"sync"
	"time"
)

type Node struct {
	Key, Value string
	UpdatedAt  time.Time
}

func newNode(key, value string) *Node {
	return &Node{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
}

type StoreMain struct {
	mutexMap sync.Map
}

func NewStoreMain() *StoreMain {
	return &StoreMain{}
}

func (s *StoreMain) Create(key, value string) error {

	n := newNode(key, value)

	if _, ok := s.mutexMap.Load(key); ok {
		return fmt.Errorf("data sudah ada, tidak boleh duplikat")
	}
	s.mutexMap.Store(key, n)

	return nil
}

func (s *StoreMain) Update(key, newval string, now time.Time) error {
	anyOldNode, exists := s.mutexMap.Load(key)
	if !exists {
		// Jika key belum ada, buat node baru dan simpan
		// s.mutexMap.Store(key, newNode(key, newValue))
		return fmt.Errorf("data tidak ada di sistem")
	}

	// 2. Lakukan type assertion ke *Node
	oldNode := anyOldNode.(*Node)

	// 3. Buat objek Node baru dengan nilai dan waktu yang diperbarui
	updatedNode := &Node{
		Key:       oldNode.Key,
		Value:     newval,
		UpdatedAt: now, // Update timestamp di sini
	}

	// 4. Ganti node lama dengan yang baru secara atomic.
	// Jika nilainya belum berubah di goroutine lain, fungsi ini sukses (return true).
	if s.mutexMap.CompareAndSwap(key, oldNode, updatedNode) {
		// return fmt.Errorf("gagal update compare swap")
		return nil
	}
	// log.Printf("pass update to err")

	return fmt.Errorf("gagal update compare swap")
}

func (s *StoreMain) Fetch(key string) *Node {

	val, ok := s.mutexMap.Load(key)
	if !ok {
		return nil
	}
	val1, _ := val.(*Node)
	// if !ok {
	// 	log.Printf("failed parsing Fetch")
	// 	return nil
	// }
	return val1
}
