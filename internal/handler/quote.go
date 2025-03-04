package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"wise-tcp/pkg/core/build"
	"wise-tcp/pkg/log"
)

func Builder() build.Builder {
	return func(_ *build.Injector) (any, error) {
		q, err := NewQuote()
		if err != nil {
			return nil, err
		}
		return q, nil
	}
}

type Quote struct {
	quoteDB []string
	mu      sync.RWMutex
	client  *http.Client
}

const quotesBatchURL = "https://zenquotes.io/api/quotes"

const randomQuoteURL = "https://zenquotes.io/api/random"

func NewQuote() (*Quote, error) {
	q := &Quote{
		client: &http.Client{Timeout: 5 * time.Second},
	}

	return q, nil
}

func (q *Quote) Init(ctx context.Context) error {
	q.loadZenQuotes(ctx)
	return nil
}

func (q *Quote) Handle(ctx context.Context, rw io.ReadWriter) error {
	quote, err := q.getQuote(ctx)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("failed to fetch quote: %w", err)
	}

	_, err = rw.Write([]byte(quote + "\n"))
	if err != nil {
		return fmt.Errorf("failed to write to connection: %w", err)
	}

	return nil
}

type zenQuote struct {
	Q string `json:"q"`
	A string `json:"a"`
}

func (q *Quote) loadZenQuotes(ctx context.Context) {
	q.mu.Lock()
	defer q.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, quotesBatchURL, nil)
	if err != nil {
		log.Error("Failed to create request:", err)
		q.quoteDB = q.fallbackQuotes()
		return
	}

	resp, err := q.client.Do(req)
	if err != nil {
		log.Error("Failed to fetch quotes:", err)
		q.quoteDB = q.fallbackQuotes()
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("Failed to close response body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Error("Non-OK HTTP status:", resp.StatusCode)
		q.quoteDB = q.fallbackQuotes()
		return
	}

	var quotes []zenQuote
	if err = json.NewDecoder(resp.Body).Decode(&quotes); err != nil {
		log.Error("Failed to decode quotes:", err)
		q.quoteDB = q.fallbackQuotes()
		return
	}

	for _, quote := range quotes {
		q.quoteDB = append(q.quoteDB, quote.Q)
	}

	if len(q.quoteDB) == 0 {
		log.Warn("No quotes loaded from API, falling back to defaults")
		q.quoteDB = q.fallbackQuotes()
	}
}

func (q *Quote) getQuote(ctx context.Context) (string, error) {
	q.mu.RLock()
	if len(q.quoteDB) > 0 {
		quote := q.randomFromList()
		q.mu.RUnlock()
		return quote, nil
	}
	q.mu.RUnlock()

	return q.randomZenOnline(ctx)
}

func (q *Quote) randomFromList() string {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.quoteDB) == 0 {
		return "No quotes available"
	}

	var l = big.NewInt(int64(len(q.quoteDB)))
	n, err := rand.Int(rand.Reader, l)
	if err != nil {
		panic(err)
	}
	return q.quoteDB[n.Int64()]
}

func (q *Quote) randomZenOnline(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, randomQuoteURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("Failed to close response body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	var quotes []zenQuote
	if err := json.NewDecoder(resp.Body).Decode(&quotes); err != nil || len(quotes) == 0 {
		return "", fmt.Errorf("failed to decode quote: %w", err)
	}

	return quotes[0].Q, nil
}

func (q *Quote) fallbackQuotes() []string {
	return []string{
		"Blessed is he who expects nothing, for he shall never be disappointed.",
		"The only real mistake is the one from which we learn nothing.",
	}
}
