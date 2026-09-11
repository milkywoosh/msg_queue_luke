package db

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

var store *StoreMain = NewStoreMain()

func TestCreate(t *testing.T) {

	err := store.Create("item01", "item01")
	require.NoError(t, err)

	// val before update
	beforeNode := store.Fetch("item01")
	nodeVal := beforeNode.Value
	dateBefore := beforeNode.UpdatedAt

	require.Equal(t, "item01", nodeVal)

	err = store.Update("item01")
	require.NoError(t, err)

	node := store.Fetch("item01")
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
