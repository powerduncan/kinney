package api

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type httpLogEntry struct {
	RequestTimestamp time.Time
	RequestMethod    string
	RequestURL       string
	RequestHeaders   map[string][]string `json:",omitempty"`
	RequestBody      []byte              `json:",omitempty"`

	ResponseTimestamp  time.Time
	ResponseStatusCode int
	ResponseHeaders    map[string][]string `json:",omitempty"`
	ResponseBody       []byte              `json:",omitempty"`

	Err error `json:",omitempty"`
}

// Issues a SOAP v1.1 request, unmarshaling the response into `respHeader` and
// `respBody`.  Details of the HTTP request and response are written to
// `httpLogWriter` as a JSON-serialized `httpLogEntry` (in JSONL format: one
// line per entry).
func soapCall(ctx context.Context, c *http.Client, url string, reqHeader, reqBody, respHeader, respBody interface{}, httpLogWriter io.Writer) error {
	reqBytes, err := marshalEnvelope(reqHeader, reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")

	logEntry := &httpLogEntry{
		RequestTimestamp: time.Now(),
		RequestMethod:    httpReq.Method,
		RequestURL:       httpReq.URL.String(),
		RequestHeaders:   httpReq.Header,
		RequestBody:      reqBytes,
	}
	defer func() {
		if b, err := json.Marshal(logEntry); err != nil {
			log.Printf("error marshaling HTTP log entry: %s", err)
		} else if _, err := httpLogWriter.Write(append(b, '\n')); err != nil {
			log.Printf("error writing HTTP log entry: %s", err)
		}
	}()

	httpResp, err := c.Do(httpReq)
	logEntry.ResponseTimestamp = time.Now()
	if httpResp != nil {
		logEntry.ResponseStatusCode = httpResp.StatusCode
	}
	if err != nil {
		logEntry.Err = fmt.Errorf("error executing HTTP request: %w", err)
		return logEntry.Err
	}
	defer httpResp.Body.Close()

	respBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		logEntry.Err = fmt.Errorf("error reading HTTP response body: %w", err)
		return logEntry.Err
	}
	logEntry.ResponseBody = respBytes

	if httpResp.StatusCode != http.StatusOK {
		logEntry.Err = fmt.Errorf("received non-200 response: %s", httpResp.Status)
		return logEntry.Err
	}

	if err := UnmarshalEnvelope(respBytes, respHeader, respBody); err != nil {
		logEntry.Err = fmt.Errorf("error unmarshaling response: %w", err)
		return logEntry.Err
	}
	return nil
}

// envelope is a struct representation of a SOAP v1.1 "Envelope" message, which
// can be marshalled and unmarshalled to/from valid SOAP requests and responses.
type envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`

	Header struct {
		Payload interface{} `xml:",any"`
	} `xml:"Header"`
	Body struct {
		Payload interface{} `xml:",any"`
	} `xml:"Body"`
}

func marshalEnvelope(header, body interface{}) ([]byte, error) {
	var env envelope
	env.Header.Payload = header
	env.Body.Payload = body

	b, err := xml.Marshal(&env)
	if err != nil {
		return nil, fmt.Errorf("error marshaling SOAP envelope: %w", err)
	}
	return b, nil
}

func UnmarshalEnvelope(b []byte, header, body interface{}) error {
	var env envelope
	env.Header.Payload = header
	env.Body.Payload = body
	if err := xml.Unmarshal(b, &env); err != nil {
		return fmt.Errorf("error unmarshaling SOAP envelope: %w", err)
	}
	return nil
}
