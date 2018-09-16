package queue

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/list"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type S struct {
	I uint64
	B bool
}

func defaultComponents(key sdk.StoreKey) (sdk.Context, *codec.Codec) {
	db := dbm.NewMemDB()
	cms := rootmulti.NewStore(db)
	cms.MountStoreWithDB(key, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := codec.New()
	return ctx, cdc
}

func TestQueue(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)

	qm := New(cdc, store)

	val := S{1, true}
	var res S

	qm.Push(val)
	qm.Peek(&res)
	require.Equal(t, val, res)

	qm.Pop()
	empty := qm.IsEmpty()

	require.True(t, empty)
	require.NotNil(t, qm.Peek(&res))

	qm.Push(S{1, true})
	qm.Push(S{2, true})
	qm.Push(S{3, true})
	qm.Flush(&res, func() (brk bool) {
		if res.I == 3 {
			brk = true
		}
		return
	})

	require.False(t, qm.IsEmpty())

	qm.Pop()
	require.True(t, qm.IsEmpty())
}

func TestKeys(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx, cdc := defaultComponents(key)
	store := ctx.KVStore(key)
	listStore := prefix.NewStore(store, ListKey())
	queue := New(cdc, store)

	for i := 0; i < 10; i++ {
		queue.Push(i)
	}

	var len uint64
	var top uint64
	var expected int
	var actual int

	// Checking keys.LengthKey
	err := cdc.UnmarshalBinary(listStore.Get(list.LengthKey()), &len)
	require.Nil(t, err)
	require.Equal(t, len, queue.List.Len())

	// Checking keys.ElemKey
	for i := 0; i < 10; i++ {
		queue.List.Get(uint64(i), &expected)
		bz := listStore.Get(list.ElemKey(uint64(i)))
		err = cdc.UnmarshalBinary(bz, &actual)
		require.Nil(t, err)
		require.Equal(t, expected, actual)
	}

	queue.Pop()

	err = cdc.UnmarshalBinary(store.Get(TopKey()), &top)
	require.Nil(t, err)
	require.Equal(t, top, queue.getTop())
}