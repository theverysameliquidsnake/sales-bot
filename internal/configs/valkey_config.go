package configs

import (
	"context"
	"fmt"
	"os"

	"github.com/valkey-io/valkey-go"
)

var client valkey.Client

func CreateValkeyClient() error {
	valk, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{os.Getenv("VALKEY_URI")}})
	if err != nil {
		return fmt.Errorf("config: could not connect to valkey: %w", err)
	}

	client = valk

	return nil
}

func DestroyValkeyClient() {
	client.Close()
}

func SetValkeyValue(key string, value string) error {
	err := client.Do(context.Background(), client.B().Set().Key(key).Value(value).Build()).Error()
	if err != nil {
		return fmt.Errorf("config: could not set valkey value: %w", err)
	}

	return nil
}

func GetValkeyValue(key string) (string, error) {
	value, err := client.Do(context.Background(), client.B().Get().Key(key).Build()).ToString()
	if err != nil {
		//return "", fmt.Errorf("config: could not get valkey value: %w", err)
		return "", nil
	}

	return value, nil
}

func FlushValkey() error {
	if err := client.Do(context.Background(), client.B().Flushall().Build()).Error(); err != nil {
		return fmt.Errorf("config: could not flush valkey values: %w", err)
	}

	return nil
}
