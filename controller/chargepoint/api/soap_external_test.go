package api_test

import (
	"log"
	"testing"

	"github.com/CamusEnergy/kinney/controller/chargepoint/api"
)

const envelopeXML = `<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/"><Header><Payload>foo</Payload></Header><Body><Payload>bar</Payload></Body></Envelope>`

func TestOnlyUnmarshalBody(t *testing.T) {
	var body string
	if err := api.UnmarshalEnvelope([]byte(envelopeXML), nil, &body); err != nil {
		t.Error(err)
	}
	log.Printf("%#v", body)
}
