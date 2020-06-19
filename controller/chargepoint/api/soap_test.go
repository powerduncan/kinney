package api

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSOAPCall(t *testing.T) {
	// Create an HTTP server for testing.  This binds to a random port on
	// the loopback interface, the address of which can be retrieved as
	// `server.URL`.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if b, err := marshalEnvelope("respHeader", "respBody"); err != nil {
			// There is no simple way to propogate this error to the
			// calling code, so just panic.
			panic(err)
		} else if _, err := w.Write(b); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	// Issue a SOAP request against the HTTP server created above, and make
	// sure that the response is umarshalled correctly.
	var respHeader, respBody string
	var httpLog bytes.Buffer
	if err := soapCall(context.Background(), &http.Client{}, server.URL, "reqHeader", "reqBody", &respHeader, &respBody, &httpLog); err != nil {
		t.Errorf("soapCall() = %q; want nil", err)
	} else if respHeader != "respHeader" || respBody != "respBody" {
		t.Errorf("respHeader = %q, respBody = %q; want %q and %q", respHeader, respBody, "respHeader", "respBody")
	}

	var logEntry httpLogEntry
	if err := json.Unmarshal(httpLog.Bytes(), &logEntry); err != nil {
		t.Errorf("json.Unmarshal(%q) = %q; want nil", httpLog.Bytes(), err)
	}

	expectedReqBody, _ := marshalEnvelope("reqHeader", "reqBody")
	expectedRespBody, _ := marshalEnvelope("respHeader", "respBody")

	// Validate the log entry that was written to the HTTP log strream.
	expectedLogEntry := httpLogEntry{
		RequestTimestamp:   time.Now(),
		RequestMethod:      "POST",
		RequestURL:         server.URL,
		RequestHeaders:     http.Header{"Content-Type": {"text/xml; charset=utf-8"}},
		RequestBody:        expectedReqBody,
		ResponseTimestamp:  time.Now(),
		ResponseStatusCode: 200,
		ResponseHeaders:    nil,
		ResponseBody:       expectedRespBody,
		Err:                nil,
	}
	if diff := cmp.Diff(&expectedLogEntry, &logEntry, cmp.Options{cmpopts.EquateApproxTime(time.Second)}); diff != "" {
		t.Errorf("log entry mismatch (-want +got):\n%s", diff)
	}
}

var envelopeXML = regexp.MustCompile(`\s*<`).ReplaceAllString(`
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
  <Header>
    <Payload xmlns="http://example.com">
      <Element>header</Element>
    </Payload>
  </Header>
  <Body>
    <Payload xmlns="http://example.com">
      <Element>body</Element>
    </Payload>
  </Body>
</Envelope>`, "<")

type payload struct {
	XMLName xml.Name `xml:"http://example.com Payload"`

	Element string `xml:"Element"`
}

func TestMarshalUnmarshal(t *testing.T) {
	header := payload{Element: "header"}
	body := payload{Element: "body"}

	b, err := marshalEnvelope(&header, &body)
	if err != nil {
		t.Errorf("marshalEnvelope(%#v, %#v) = %q; want nil", &header, &body, err)
	} else if diff := cmp.Diff(envelopeXML, string(b)); diff != "" {
		t.Errorf("marshalEnvelope(%#v, %#v) mismatch (-want +got):\n%s", &header, &body, diff)
	}

	var parsedHeader, parsedBody payload
	if err := UnmarshalEnvelope(b, &parsedHeader, &parsedBody); err != nil {
		t.Errorf("UnmarshalEnvelope(%q) = %q; want nil", string(b), err)
	}

	// Set the `XMLName` field in the original values since it will be set
	// in the unmarshaled values (it will be set to the full name of the
	// element that were unmarshaled into the value).
	header.XMLName = xml.Name{"http://example.com", "Payload"}
	body.XMLName = xml.Name{"http://example.com", "Payload"}
	if diff := cmp.Diff(header, parsedHeader); diff != "" {
		t.Errorf("UnmarshalEnvelope(%q) mismatch (-want +got):\n%s", string(b), diff)
	} else if diff := cmp.Diff(body, parsedBody); diff != "" {
		t.Errorf("UnmarshalEnvelope(%q) mismatch (-want +got):\n%s", string(b), diff)
	}
}
