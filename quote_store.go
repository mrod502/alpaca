package alpaca

import (
	"errors"
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/v2/stream"
	"go.uber.org/atomic"
)

var (
	ErrIndex = errors.New("index out of bounds")
)

func NewQuoteStore(cap uint32, data ...stream.Quote) *QuoteStore {
	q := &QuoteStore{
		start: atomic.NewUint32(0),
		cap:   maxu32(uint32(len(data)), cap),
		size:  atomic.NewUint32(0),
	}

	q.d = make([]stream.Quote, q.cap)
	for k, v := range data {
		q.d[k] = v
	}
	return q
}

type QuoteStore struct {
	start *atomic.Uint32
	d     []stream.Quote
	cap   uint32
	size  *atomic.Uint32
}

func (q *QuoteStore) String() string {
	return fmt.Sprintf("{start:%d,cap:%d,size:%d}", q.start.Load(), q.cap, q.size.Load())
}

func (q *QuoteStore) Write(v stream.Quote) {
	if sz := q.size.Load(); sz < q.cap {
		q.d[sz] = v
		q.size.Inc()
		return
	}
	st := q.start.Load()
	q.d[st] = v
	q.start.Store((st + 1) % q.cap)
}

func (q *QuoteStore) Last() stream.Quote {
	if q.size.Load() == 0 {
		return stream.Quote{}
	}
	return q.d[q.getIx(q.size.Load()-1)]
}

func (q *QuoteStore) ValueAt(ix uint32) (stream.Quote, error) {
	if ix >= q.cap {
		return stream.Quote{}, ErrIndex
	}
	return q.d[q.getIx(ix)], nil
}

func (q *QuoteStore) Slice(begin, end uint32) ([]stream.Quote, error) {
	if (begin >= q.cap) || (end >= q.cap) {
		return nil, ErrIndex
	}
	if begin > end {
		begin, end = end, begin
		return reverse(q.sliceBase(begin, end)...), nil
	}
	return q.sliceBase(begin, end), nil
}

func (q *QuoteStore) sliceBase(begin, end uint32) []stream.Quote {
	if q.getIx(begin) <= q.getIx(end) {
		return q.d[q.getIx(begin):q.getIx(end)]
	}
	return append(q.d[q.getIx(begin):], q.d[0:q.getIx(end)]...)
}

func (q *QuoteStore) getIx(ix uint32) uint32 {
	return (q.start.Load() + ix) % q.cap
}

func maxu32(vals ...uint32) (m uint32) {
	for _, v := range vals {
		if v > m {
			m = v
		}
	}
	return
}

func reverse(vals ...stream.Quote) []stream.Quote {
	for i, j := 0, len(vals)-1; i < j; i, j = i+1, j-1 {
		vals[j], vals[i] = vals[i], vals[j]
	}
	return vals
}
