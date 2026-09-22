// Package audit defines the shared record format written by the producer
// Lambda (HTTP -> SQS) and consumed by the consumer Lambda (SQS -> S3).
// Keeping this in one place means both functions can never drift apart
// on the schema.
package audit

// Record is the full archived entry for a single API request.
type Record struct {
	RequestID      string            `json:"request_id"`
	Timestamp      string            `json:"timestamp"`
	Path           string            `json:"path"`
	HTTPMethod     string            `json:"http_method"`
	Headers        map[string]string `json:"headers"`
	QueryParams    map[string]string `json:"query_params"`
	RequestBody    string            `json:"request_body"`
	ResponseStatus int               `json:"response_status"`
	ResponseBody   string            `json:"response_body"`
}
