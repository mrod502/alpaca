package alpaca

import (
	"errors"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/alpacahq/alpaca-trade-api-go/v2/stream"
	"github.com/google/uuid"
	gocache "github.com/mrod502/go-cache"
	"go.uber.org/atomic"
)

var (
	ErrTooManySymbols = errors.New("too many symbols subscribed")
)

type ApiKey struct {
	Id     string
	Secret string
}

type Client struct {
	cli         *alpaca.Client
	quotes      *gocache.InterfaceCache
	subscribers *gocache.InterfaceCache
	numSymbols  *atomic.Uint32
	maxSymbols  uint32
	numQuotes   uint32
}

func NewClient(key ApiKey, ttl time.Duration, numQuotes uint32) *Client {
	os.Setenv(common.EnvApiKeyID, key.Id)
	os.Setenv(common.EnvApiSecretKey, key.Secret)

	return &Client{
		cli:         alpaca.NewClient(common.Credentials()),
		quotes:      gocache.NewInterfaceCache(),
		subscribers: gocache.NewInterfaceCache(),
		numSymbols:  atomic.NewUint32(0),
		maxSymbols:  30,
		numQuotes:   numQuotes,
	}
}

func (c *Client) SubscribeQuotes(s string) error {
	if c.numSymbols.Load() >= c.maxSymbols {
		return ErrTooManySymbols
	}
	return stream.SubscribeQuotes(c.updateQuotes, s)
}

func (c *Client) updateQuotes(q stream.Quote) {
	qs := c.getQuoteStore(q.Symbol)
	qs.Write(q)
	c.dispatch(qs)
}

func (c *Client) getQuoteStore(k string) *QuoteStore {
	if !c.quotes.Exists(k) {
		c.quotes.Set(k, NewQuoteStore(c.numQuotes))
	}
	return c.quotes.Get(k).(*QuoteStore)
}

func (c *Client) UnsubscribeQuotes(s string) error {
	if !c.quotes.Exists(s) {
		return nil
	}
	defer c.quotes.Delete(s)
	return stream.UnsubscribeQuotes(s)
}

func (c *Client) AddSubscriber(ch chan interface{}) string {
	uid := uuid.New().String()
	c.subscribers.Set(uid, ch)
	return uid
}

func (c *Client) RemoveSubscriber(k string) error {
	return c.subscribers.Delete(k)
}

func (c *Client) dispatch(v interface{}) {
	for _, k := range c.subscribers.GetKeys() {
		ch, ok := c.getSubscriber(k)
		if !ok || len(ch) == cap(ch) {
			continue
		}
		ch <- v
	}
}

func (c *Client) getSubscriber(k string) (v chan interface{}, ok bool) {
	v, ok = c.subscribers.Get(k).(chan interface{})
	return
}
