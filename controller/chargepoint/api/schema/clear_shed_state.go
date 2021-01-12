package schema

import "encoding/xml"

// API Guide (§ 6.2): "Use this call to clear the shed state from a single
// station or group of stations."
type ClearShedStateRequest struct {
	XMLName xml.Name `xml:"urn:dictionary:com.chargepoint.webservices clearShedState"`

	StationGroupID *int32 `xml:"shedQuery>shedGroup>sgID,omitempty"`

	StationID *string `xml:"shedQuery>shedStation>stationID,omitempty"`

	// This part of the API is not documented in the API Guide, but it is
	// part of the WSDL.
	//
	// Note that `StationID` is required if `PortNumbers` is specified.
	//
	// TODO(james): Check if this is actually implemented in the API server.
	// If not, it should be removed from this schema representation.
	Ports *ClearShedStateRequest_Ports `xml:"shedQuery>shedStation>Ports,omitempty"`
}

type ClearShedStateRequest_Ports struct {
	PortNumbers []string `xml:"Port>portNumber"`
}

type ClearShedStateResponse struct {
	XMLName xml.Name `xml:"urn:dictionary:com.chargepoint.webservices clearShedStateResponse"`

	commonResponseParameters

	// API Guide (§ 6.2.3): "A success (1) or failure (0) response code
	// only."
	//
	// This field has `type="xsd:int"` in the WSDL, but is semantically
	// boolean, and can safely be parsed as such.
	Success bool `xml:"Success,omitempty"`

	StationGroupID int32  `xml:"sgID,omitempty"`
	StationID      string `xml:"stationID,omitempty"`
}
