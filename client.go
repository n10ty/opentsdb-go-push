package opentsdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const defaultBatchSize = 20

// Client has 2 options to send metrics:
// - Enqueue metrics, send when batchSize collected and flush buffer. Use Push to force send current buffer.
// - Send single Metric immediately.
type Client struct {
	url       string
	authUser  string
	authPass  string
	buffer    []Metric
	batchSize int
}

type Metric struct {
	Timestamp int64             `json:"timestamp"`
	Metric    string            `json:"metric"`
	Value     any               `json:"value"`
	Tags      map[string]string `json:"tags"`
}

type config struct {
	authUsername string
	authPassword string
	batchSize    int
}

func NewClient(url string, options ...Option) (*Client, error) {
	config := &config{
		authUsername: "",
		authPassword: "",
		batchSize:    defaultBatchSize,
	}
	for _, o := range options {
		err := o(config)
		if err != nil {
			return nil, fmt.Errorf("failed to construc opentsdb client: %w", err)
		}
	}

	return &Client{
		url:       url,
		authUser:  config.authUsername,
		authPass:  config.authPassword,
		batchSize: config.batchSize,
	}, nil
}

// Enqueue send metric to a buffer. Metrics are sent when buffer reaches batchSize number.
func (c *Client) Enqueue(metric Metric) error {
	if metric.Tags == nil {
		return errors.New("tags can not be nil")
	}
	c.buffer = append(c.buffer, metric)
	if len(c.buffer) >= c.batchSize {
		err := c.send(c.buffer)
		c.buffer = []Metric{}
		if err != nil {
			return err
		}
	}
	return nil
}

// Send single Metric immediately
func (c *Client) Send(metric Metric) error {
	if metric.Tags == nil {
		return errors.New("tags can not be nil")
	}
	return c.send([]Metric{metric})
}

func (c *Client) send(metric []Metric) error {
	url := fmt.Sprintf("%s/api/put", c.url)
	m, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, body(m))
	if err != nil {
		return err
	}

	if c.authUser != "" {
		req.SetBasicAuth(c.authUser, c.authPass)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: %s", res.Status, string(b))
	}

	return nil
}

// Push buffer and clean it
func (c *Client) Push() error {
	if len(c.buffer) == 0 {
		return nil
	}
	err := c.send(c.buffer)
	c.buffer = []Metric{}
	if err != nil {
		return err
	}
	return nil
}

// Close should be used on service down to prevent an unfilled buffer to be gone
func (c *Client) Close() error {
	return c.send(c.buffer)
}

func body(buf []byte) io.Reader {
	return bytes.NewBuffer(buf)
}

type Option func(*config) error

// WithAuth setup BasicAuth username and password to use in request
func WithAuth(username, password string) Option {
	return func(c *config) error {
		c.authUsername = username
		c.authPassword = password
		return nil
	}
}

// WithBatchSize change default number of buffer size to push
func WithBatchSize(n int) Option {
	return func(c *config) error {
		if n < 1 || n > 1024 {
			return errors.New("batch size should be between 1 and 1024")
		}
		c.batchSize = n
		return nil
	}
}
