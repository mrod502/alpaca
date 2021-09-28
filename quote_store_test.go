package alpaca

import (
	"fmt"
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/v2/stream"
)

func TestQuoteStore(t *testing.T) {
	s := NewQuoteStore(5)

	for i := 0; i < 10; i++ {
		s.Write(stream.Quote{Symbol: "foo", AskPrice: float64(i)})
		v, _ := s.ValueAt(0)
		fmt.Println(v.AskPrice)
	}
}
