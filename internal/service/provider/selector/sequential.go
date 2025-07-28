package selector

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/service/provider"
)

var _ provider.Selector = (*SeqSelector)(nil)

type SeqSelector struct {
	index     int
	providers []provider.Provider
}

func (s *SeqSelector) Next(_ context.Context, _ domain.Notification) (provider.Provider, error) {
	if len(s.providers) == s.index {
		return nil, fmt.Errorf("%w: no available provider", errs.ErrRecordNotFound)
	}
	p := s.providers[s.index]
	s.index++
	return p, nil
}

var _ provider.SelectorBuilder = (*SeqSelectorBuilder)(nil)

type SeqSelectorBuilder struct {
	providers []provider.Provider
}

func (b *SeqSelectorBuilder) Build() (provider.Selector, error) {
	return &SeqSelector{
		providers: b.providers,
	}, nil
}

func NewSeqSelectorBuilder(providers []provider.Provider) *SeqSelectorBuilder {
	return &SeqSelectorBuilder{
		providers: providers,
	}
}
