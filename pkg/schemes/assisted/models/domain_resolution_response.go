// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// DomainResolutionResponse domain resolution response
//
// swagger:model domain_resolution_response
type DomainResolutionResponse struct {

	// resolutions
	// Required: true
	Resolutions []*DomainResolutionResponseDomain `json:"resolutions"`
}

// Validate validates this domain resolution response
func (m *DomainResolutionResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateResolutions(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *DomainResolutionResponse) validateResolutions(formats strfmt.Registry) error {

	if err := validate.Required("resolutions", "body", m.Resolutions); err != nil {
		return err
	}

	for i := 0; i < len(m.Resolutions); i++ {
		if swag.IsZero(m.Resolutions[i]) { // not required
			continue
		}

		if m.Resolutions[i] != nil {
			if err := m.Resolutions[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("resolutions" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("resolutions" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this domain resolution response based on the context it is used
func (m *DomainResolutionResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateResolutions(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *DomainResolutionResponse) contextValidateResolutions(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Resolutions); i++ {

		if m.Resolutions[i] != nil {
			if err := m.Resolutions[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("resolutions" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("resolutions" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *DomainResolutionResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *DomainResolutionResponse) UnmarshalBinary(b []byte) error {
	var res DomainResolutionResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// DomainResolutionResponseDomain domain resolution response domain
//
// swagger:model DomainResolutionResponseDomain
type DomainResolutionResponseDomain struct {

	// The cnames that were resolved for the domain, empty if none
	Cnames []string `json:"cnames"`

	// The domain that was resolved
	// Required: true
	DomainName *string `json:"domain_name"`

	// The IPv4 addresses of the domain, empty if none
	IPV4Addresses []strfmt.IPv4 `json:"ipv4_addresses"`

	// The IPv6 addresses of the domain, empty if none
	IPV6Addresses []strfmt.IPv6 `json:"ipv6_addresses"`
}

// Validate validates this domain resolution response domain
func (m *DomainResolutionResponseDomain) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDomainName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIPV4Addresses(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIPV6Addresses(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *DomainResolutionResponseDomain) validateDomainName(formats strfmt.Registry) error {

	if err := validate.Required("domain_name", "body", m.DomainName); err != nil {
		return err
	}

	return nil
}

func (m *DomainResolutionResponseDomain) validateIPV4Addresses(formats strfmt.Registry) error {
	if swag.IsZero(m.IPV4Addresses) { // not required
		return nil
	}

	for i := 0; i < len(m.IPV4Addresses); i++ {

		if err := validate.FormatOf("ipv4_addresses"+"."+strconv.Itoa(i), "body", "ipv4", m.IPV4Addresses[i].String(), formats); err != nil {
			return err
		}

	}

	return nil
}

func (m *DomainResolutionResponseDomain) validateIPV6Addresses(formats strfmt.Registry) error {
	if swag.IsZero(m.IPV6Addresses) { // not required
		return nil
	}

	for i := 0; i < len(m.IPV6Addresses); i++ {

		if err := validate.FormatOf("ipv6_addresses"+"."+strconv.Itoa(i), "body", "ipv6", m.IPV6Addresses[i].String(), formats); err != nil {
			return err
		}

	}

	return nil
}

// ContextValidate validates this domain resolution response domain based on context it is used
func (m *DomainResolutionResponseDomain) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *DomainResolutionResponseDomain) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *DomainResolutionResponseDomain) UnmarshalBinary(b []byte) error {
	var res DomainResolutionResponseDomain
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}