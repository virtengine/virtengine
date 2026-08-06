package documents

import (
	"context"
	"fmt"
	"image"
)

type Registry struct {
	adapters []DocumentAdapter
}

func NewRegistry(adapters ...DocumentAdapter) *Registry {
	reg := &Registry{}
	for _, adapter := range adapters {
		reg.Register(adapter)
	}
	return reg
}

func (r *Registry) Register(adapter DocumentAdapter) {
	if adapter == nil {
		return
	}
	r.adapters = append(r.adapters, adapter)
}

func (r *Registry) AdapterFor(docType DocumentType, country CountryCode) (DocumentAdapter, bool) {
	for _, adapter := range r.adapters {
		if adapter.CanProcess(docType, country) {
			return adapter, true
		}
	}
	return nil, false
}

func (r *Registry) Extract(ctx context.Context, docType DocumentType, country CountryCode, img image.Image, mrzValue string) (*DocumentData, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, ok := r.AdapterFor(docType, country)
	if !ok {
		return nil, ErrNoAdapter
	}
	var data *DocumentData
	var err error
	if mrzValue != "" {
		data, err = adapter.ExtractWithMRZ(ctx, img, mrzValue)
	} else {
		data, err = adapter.Extract(ctx, img)
	}
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("%w: adapter returned no document data", ErrInvalidDocument)
	}
	if data.DocumentType != docType {
		return nil, fmt.Errorf("%w: requested type %s, extracted %s", ErrInvalidDocument, docType, data.DocumentType)
	}
	if data.IssuingCountry != country {
		return nil, fmt.Errorf("%w: requested country %s, extracted %s", ErrInvalidDocument, country, data.IssuingCountry)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	validationErrors, err := adapter.Validate(data)
	if err != nil {
		return nil, fmt.Errorf("%w: adapter validation failed: %v", ErrInvalidDocument, err)
	}
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("%w: adapter returned validation errors", ErrInvalidDocument)
	}
	return data, nil
}
