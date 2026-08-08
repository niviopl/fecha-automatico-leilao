package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestCreateAuction_ClosesAutomaticallyAfterInterval proves that an auction
// created with a status of Active transitions on its own to Closed once
// AUCTION_INTERVAL elapses, with no manual intervention.
//
// It requires a reachable MongoDB (e.g. `docker-compose up -d mongodb`);
// it is skipped automatically when none is available.
func TestCreateAuction_ClosesAutomaticallyAfterInterval(t *testing.T) {
	const auctionInterval = 2 * time.Second
	os.Setenv("AUCTION_INTERVAL", auctionInterval.String())
	defer os.Unsetenv("AUCTION_INTERVAL")

	mongoURL := os.Getenv("TEST_MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/auctions?authSource=admin"
	}

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelConnect()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(mongoURL))
	require.NoError(t, err)
	defer client.Disconnect(context.Background())

	if err := client.Ping(connectCtx, nil); err != nil {
		t.Skipf("mongodb not reachable at %s, skipping integration test: %v", mongoURL, err)
	}

	database := client.Database("auctions_test")
	repo := NewAuctionRepository(database)
	defer database.Collection("auctions").Drop(context.Background())

	auction, err := auction_entity.CreateAuction(
		"Product Test",
		"Category Test",
		"Description with more than ten characters",
		auction_entity.New)
	require.Nil(t, err)

	createErr := repo.CreateAuction(context.Background(), auction)
	require.Nil(t, createErr)

	// 1. Auction is created open.
	created, findErr := repo.FindAuctionById(context.Background(), auction.Id)
	require.Nil(t, findErr)
	assert.Equal(t, auction_entity.Active, created.Status)

	// 2. Wait for the configured AUCTION_INTERVAL to elapse.
	time.Sleep(auctionInterval + 2*time.Second)

	// 3. Status must have flipped to Closed automatically.
	closed, findErr := repo.FindAuctionById(context.Background(), auction.Id)
	require.Nil(t, findErr)
	assert.Equal(t, auction_entity.Completed, closed.Status)
}
