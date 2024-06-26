//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/common"
)

// CertificateLocations shall represent the certificate location properties for a Redfish implementation.
type CertificateLocations struct {
	common.Entity
	// ODataContext is the odata context.
	ODataContext string `json:"@odata.context"`
	// ODataEtag is the odata etag.
	ODataEtag string `json:"@odata.etag"`
	// ODataType is the odata type.
	ODataType string `json:"@odata.type"`
	// Description provides a description of this resource.
	Description string
	// certificates is the link to the certificates installed on this service.
	certificates []string
	// CertificatesCount is the number of certificates installed on this service.
	CertificatesCount int
}

// UnmarshalJSON unmarshals a CertificateLocations object from the raw JSON.
func (certificatelocations *CertificateLocations) UnmarshalJSON(b []byte) error {
	type temp CertificateLocations
	type Links struct {
		// Certificates shall contain an array of links to resources of type Certificate that are installed on this
		// service.
		Certificates common.Links
		// Certificates@odata.count
		CertificatesCount int `json:"Certificates@odata.count"`
	}
	var t struct {
		temp
		Links Links
	}

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	*certificatelocations = CertificateLocations(t.temp)

	// Extract the links to other entities for later
	certificatelocations.certificates = t.Links.Certificates.ToStrings()
	certificatelocations.CertificatesCount = t.Links.CertificatesCount

	return nil
}

// GetCertificateLocations will get a CertificateLocations instance from the service.
func GetCertificateLocations(c common.Client, uri string) (*CertificateLocations, error) {
	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var certificatelocations CertificateLocations
	err = json.NewDecoder(resp.Body).Decode(&certificatelocations)
	if err != nil {
		return nil, err
	}

	certificatelocations.SetClient(c)
	return &certificatelocations, nil
}

// ListReferencedCertificateLocationss gets the collection of CertificateLocations from
// a provided reference.
func ListReferencedCertificateLocations(c common.Client, link string) ([]*CertificateLocations, error) {
	var result []*CertificateLocations
	if link == "" {
		return result, nil
	}

	type GetResult struct {
		Item  *CertificateLocations
		Link  string
		Error error
	}

	ch := make(chan GetResult)
	collectionError := common.NewCollectionError()
	get := func(link string) {
		certificatelocations, err := GetCertificateLocations(c, link)
		ch <- GetResult{Item: certificatelocations, Link: link, Error: err}
	}

	go func() {
		err := common.CollectList(get, c, link)
		if err != nil {
			collectionError.Failures[link] = err
		}
		close(ch)
	}()

	for r := range ch {
		if r.Error != nil {
			collectionError.Failures[r.Link] = r.Error
		} else {
			result = append(result, r.Item)
		}
	}

	if collectionError.Empty() {
		return result, nil
	}

	return result, collectionError
}

// Certificates retrieves a collection of the Certificates installed on the system.
func (certificatelocations *CertificateLocations) Certificates() ([]*Certificate, error) {
	var result []*Certificate

	collectionError := common.NewCollectionError()
	for _, uri := range certificatelocations.certificates {
		cert, err := GetCertificate(certificatelocations.GetClient(), uri)
		if err != nil {
			collectionError.Failures[uri] = err
		} else {
			result = append(result, cert)
		}
	}

	if collectionError.Empty() {
		return result, nil
	}

	return result, collectionError
}
