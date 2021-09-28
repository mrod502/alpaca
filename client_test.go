package alpaca

import (
	"encoding/json"
	"os"
	"path"
	"testing"
	"time"

	"github.com/mrod502/logger"
)

func TestClientWebsocketFeed(t *testing.T) {
	logger.Info("WS", "starting test")
	p, _ := os.UserHomeDir()
	b, err := os.ReadFile(path.Join(p, "alpaca.json"))
	var key ApiKey

	if err != nil {
		t.Fatal(err)
	}

	json.Unmarshal(b, &key)
	logger.Info("WS", "creating client")
	cli := NewClient(key, time.Second, 100)

	err = cli.SubscribeQuotes("SPY")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Second)
}
