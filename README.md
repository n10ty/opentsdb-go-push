# OPENTSDB metrics push GO client

Client has 2 options to send metrics:
 - Enqueue metrics, send when batchSize collected and flush buffer. Use Push to force send current buffer.
 - Send single Metric immediately.

*Important*

Do not forger invoke `Close` on service down, to send unfilled buffer

## Usage

Import:

`go get github.com/n10ty/opentsdb-go-push`

Example:

```go
func main() {
    client := opentsdb.NewClient("http://localhost:4242", opentsdb.WithBatchSize(30))
	err := client.Enqueue(opentsdb.Metric{
		Timestamp: time.Now().Truncate(time.Minute).Unix(),
		Metric:    "http_response_time",
		Value:     39.3,
		Tags: map[string]string{
			"server": "server1",
		},
	})
	if err != nil {
		//...
	}
	client.Push()
}
```