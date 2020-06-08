package schema

import "encoding/xml"

// API Guide (§ 8.3): "Use this call to retrieve custom station groups for any
// organization.  It returns an array of groups for a given organization and
// lists the stations included in each group."
type GetStationGroupsRequest struct {
	XMLName xml.Name `xml:"urn:dictionary:com.chargepoint.webservices getStationGroups"`

	OrganizationID string `xml:"orgID"`
}

type GetStationGroupsResponse struct {
	XMLName xml.Name `xml:"urn:dictionary:com.chargepoint.webservices getStationGroupsResponse"`

	commonResponseParameters

	StationGroups []struct {
		OrganizationID   string `xml:"orgID,omitempty"`
		OrganizationName string `xml:"organizationName,omitempty"`

		StationGroupID   int32  `xml:"sgID,omitempty"`
		StationGroupName string `xml:"sgName,omitempty"`

		Stations []struct {
			StationID  string      `xml:"stationID,omitempty"`
			Coordinate *Coordinate `xml:"Geo,omitempty"`
		} `xml:"stationData,omitempty"`
	} `xml:"groupData,omitempty"`
}
