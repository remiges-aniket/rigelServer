// Package etcd provides an implementation of the Storage interface defined in the Rigel project.
// It uses the etcd.
package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/remiges-aniket/types"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const dialTimeout = 5 * time.Second

// EtcdStorage is implements Rigel's Storage interface using etcd v3 client.
type EtcdStorage struct {
	Client *clientv3.Client
}

var _ types.Storage = &EtcdStorage{}

// NewEtcdStorage creates a new instance of EtcdStorage using the provided endpoints
// with default settings from the package. If an optional clientv3.Config is supplied,
// it is used to configure the etcd client, overriding the default settings.
func NewEtcdStorage(endpoints []string, config ...clientv3.Config) (*EtcdStorage, error) {
	var cfg clientv3.Config
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: dialTimeout,
		}
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	storage := &EtcdStorage{Client: cli}

	// Create a context with a timeout for the status check.
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	// Perform a status check to ensure we can connect to the etcd server.
	// The 'StatusCheck' confirms the client is not only initialized but also functionally connected to the etcd cluster.
	if err := storage.StatusCheck(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	return storage, nil
}

// StatusCheck checks the status of the etcd client.
// If the function succeeds, we can assume that the connection to the etcd server is working.
func (e *EtcdStorage) StatusCheck(ctx context.Context) error {
	_, err := e.Client.MemberList(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to etcd: %w", err)
	}

	return nil
}

// Get retrieves a value from etcd based on the provided key.
// The function returns the corresponding value as a string. If the key does not exist in etcd,
// the function returns an empty string and no error. If an error occurs during the operation,
// it is returned by the function.
func (e *EtcdStorage) Get(ctx context.Context, key string) (string, error) {
	resp, err := e.Client.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key from etcd: %w", err)
	}

	// Assuming the value is a string
	var value string
	for _, ev := range resp.Kvs {
		value = string(ev.Value)
	}

	return value, nil
}

// example ; GetWithParams(c, keyStr, clientv3.WithPrefix())
func (e *EtcdStorage) GetWithPrefix(ctx context.Context, prefix string) (map[string]string, error) {

	resp, err := e.Client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from etcd: %w", err)
	}
	keyVal := make(map[string]string)
	for _, ev := range resp.Kvs {
		keyVal[string(ev.Key)] = string(ev.Value)
	}

	return keyVal, nil
}

// Put stores a value in etcd at the specified key.
// The value is also stored as a string. If the key already exists in etcd,
// its value is updated with the new value. If the key does not exist,
// a new key-value pair is created in etcd. If an error occurs during the operation,
// it is returned by the function.
func (e *EtcdStorage) Put(ctx context.Context, key string, value string) error {
	_, err := e.Client.Put(ctx, key, value)
	if err != nil {
		return err
	}
	return nil
}

// Watch starts watching for changes to a key or a range of keys in etcd and sends the events to the provided channel.
// If the key is a prefix that matches multiple keys, it watches all those keys.
// key: The key to watch for changes
// events is the channel to send events when the key's value changes
func (e *EtcdStorage) Watch(ctx context.Context, key string, events chan<- types.Event) error {
	watchChan := e.Client.Watch(ctx, key, clientv3.WithPrefix())
	go func() {
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				events <- types.Event{
					Key:   string(event.Kv.Key),
					Value: string(event.Kv.Value),
				}
			}
		}
	}()
	return nil
}
