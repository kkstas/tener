package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/kkstas/tjener/internal/database"
)

const testTimeout = 10 * time.Second

func TestDDBTableExists(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client, err := database.CreateLocalDynamoDBClient(ctx)
	if err != nil {
		t.Fatalf("creating local dynamodb client failed, %v", err)
	}

	exists, err := database.DDBTableExists(ctx, client, "asdf1234")
	if err != nil {
		t.Fatalf("checking if local dynamodb table exists failed, %v", err)
	}

	if exists {
		t.Error("expected to return false for non-existent table name")
	}
}
