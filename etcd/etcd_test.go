package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/remiges-aniket/types"
	"go.etcd.io/etcd/tests/v3/integration"
)

func TestGetNonExistentKey(t *testing.T) {
	// Setup the test environment
	integration.BeforeTestExternal(t)

	// Create an embedded etcd server for testing
	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	// Create an EtcdStorage instance
	etcdStorage := &EtcdStorage{
		Client: clus.RandClient(),
	}

	// Try to get a non-existent key
	value, err := etcdStorage.Get(context.Background(), "non-existent-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the returned value is an empty string
	if value != "" {
		t.Errorf("Expected an empty string, got '%s'", value)
	}
}

func TestEtcdStorage_Watch(t *testing.T) {
	// Setup the test environment
	integration.BeforeTestExternal(t)

	// Create an embedded etcd server for testing
	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	// Create an EtcdStorage instance
	etcdStorage := &EtcdStorage{
		Client: clus.RandClient(),
	}

	// Create a channel for events
	events := make(chan types.Event)

	// Create a channel for errors
	errs := make(chan error)

	// Start watching a key
	go func() {
		err := etcdStorage.Watch(context.Background(), "test-key", events)
		if err != nil {
			errs <- err
		}
	}()

	// Check for any errors
	select {
	case err := <-errs:
		t.Fatalf("Expected no error, got %v", err)
	default:
	}

	// Put a value to the key
	err := etcdStorage.Put(context.Background(), "test-key", "test-value")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the event is received
	select {
	case event := <-events:
		if event.Key != "test-key" || event.Value != "test-value" {
			t.Errorf("Expected event with key 'test-key' and value 'test-value', got key '%s' and value '%s'", event.Key, event.Value)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Expected to receive an event, but didn't")
	}
}
