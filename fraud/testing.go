package fraud

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/header"
)

type DummyService struct {
}

func (d *DummyService) Broadcast(context.Context, Proof) error {
	return nil
}

func (d *DummyService) Subscribe(ProofType) (Subscription, error) {
	return &subscription{}, nil
}

func (d *DummyService) Get(context.Context, ProofType) ([]Proof, error) {
	return nil, nil
}

type mockStore struct {
	headers    map[int64]*header.ExtendedHeader
	headHeight int64
}

// createStore creates a mock store and adds several random
// headers.
func createStore(t *testing.T, numHeaders int) *mockStore {
	store := &mockStore{
		headers:    make(map[int64]*header.ExtendedHeader),
		headHeight: 0,
	}

	suite := header.NewTestSuite(t, numHeaders)

	for i := 0; i < numHeaders; i++ {
		header := suite.GenExtendedHeader()
		store.headers[header.Height] = header

		if header.Height > store.headHeight {
			store.headHeight = header.Height
		}
	}
	return store
}

func (m *mockStore) GetByHeight(_ context.Context, height uint64) (*header.ExtendedHeader, error) {
	return m.headers[int64(height)], nil
}

func (m *mockStore) Close() error { return nil }

const (
	mockProofType ProofType = "mockProof"
)

type mockProof struct {
	Valid bool
}

func newValidProof() *mockProof {
	return newMockProof(true)
}

func newInvalidProof() *mockProof {
	return newMockProof(false)
}

func newMockProof(valid bool) *mockProof {
	p := &mockProof{valid}
	if _, ok := defaultUnmarshalers[p.Type()]; !ok {
		Register(&mockProof{})
	}
	return p
}

func (m *mockProof) Type() ProofType {
	return mockProofType
}

func (m *mockProof) HeaderHash() []byte {
	return []byte("hash")
}

func (m *mockProof) Height() uint64 {
	return 1
}

func (m *mockProof) Validate(*header.ExtendedHeader) error {
	if !m.Valid {
		return errors.New("mockProof: proof is not valid")
	}
	return nil
}

func (m *mockProof) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m *mockProof) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

func CreateTestService(t *testing.T, enabledSyncer bool) (*ProofService, *mockStore) { //nolint:revive
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	t.Cleanup(cancel)

	// create mock network
	net, err := mocknet.FullMeshLinked(1)
	require.NoError(t, err)
	return createTestServiceWithHost(ctx, t, net.Hosts()[0], enabledSyncer)
}

func createTestServiceWithHost(
	ctx context.Context,
	t *testing.T,
	host host.Host,
	enabledSyncer bool,
) (*ProofService, *mockStore) {
	// create pubsub for host
	ps, err := pubsub.NewGossipSub(ctx, host,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign))
	require.NoError(t, err)
	store := createStore(t, 10)
	return NewProofService(
		ps,
		host,
		store.GetByHeight,
		sync.MutexWrap(datastore.NewMapDatastore()),
		enabledSyncer,
		"private",
	), store
}
