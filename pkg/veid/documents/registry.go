package documents

import (
	"context"
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
	adapter, ok := r.AdapterFor(docType, country)
	if !ok {
		return nil, ErrNoAdapter
	}
	if mrzValue != "" {
		return adapter.ExtractWithMRZ(ctx, img, mrzValue)
	}
	return adapter.Extract(ctx, img)
}
