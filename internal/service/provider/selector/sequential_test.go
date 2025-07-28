package selector

import (
	"errors"
	"testing"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/service/provider"
	providermock "github.com/JrMarcco/kuryr/internal/service/provider/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewSeqSelectorBuilder(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name      string
		providers func(*gomock.Controller) []provider.Provider
	}{
		{
			name: "empty providers",
			providers: func(_ *gomock.Controller) []provider.Provider {
				return []provider.Provider{}
			},
		}, {
			name: "single provider",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
				}
			},
		}, {
			name: "multiple providers",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
					providermock.NewMockProvider(ctrl),
					providermock.NewMockProvider(ctrl),
				}
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			builder := NewSeqSelectorBuilder(tc.providers(ctrl))
			assert.NotNil(t, builder)
		})
	}
}

func TestSeqSelectorBuilder_Build(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name      string
		providers func(*gomock.Controller) []provider.Provider
	}{
		{
			name: "empty providers",
			providers: func(_ *gomock.Controller) []provider.Provider {
				return []provider.Provider{}
			},
		}, {
			name: "single provider",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
				}
			},
		}, {
			name: "multiple providers",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
				}
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			providers := tc.providers(ctrl)
			builder := NewSeqSelectorBuilder(providers)

			selector, err := builder.Build()
			assert.NoError(t, err)
			assert.NotNil(t, selector)
		})
	}
}

func TestSeqSelector_Next(t *testing.T) {
	t.Parallel()

	n := domain.Notification{}

	tcs := []struct {
		name      string
		providers func(*gomock.Controller) []provider.Provider
		callTimes int
		wantErr   error
	}{
		{
			name: "empty providers",
			providers: func(_ *gomock.Controller) []provider.Provider {
				return []provider.Provider{}
			},
			callTimes: 1,
			wantErr:   errs.ErrRecordNotFound,
		}, {
			name: "single provider with one	call",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
				}
			},
			callTimes: 1,
			wantErr:   nil,
		}, {
			name: "single provider with multiple calls",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
				}
			},
			callTimes: 3,
			wantErr:   errs.ErrRecordNotFound,
		}, {
			name: "multiple providers with all get",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
					providermock.NewMockProvider(ctrl),
					providermock.NewMockProvider(ctrl)}
			},
			callTimes: 3,
			wantErr:   nil,
		}, {
			name: "multiple providers with over get",
			providers: func(ctrl *gomock.Controller) []provider.Provider {
				return []provider.Provider{
					providermock.NewMockProvider(ctrl),
					providermock.NewMockProvider(ctrl),
					providermock.NewMockProvider(ctrl),
				}
			},
			callTimes: 4,
			wantErr:   errs.ErrRecordNotFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			providers := tc.providers(ctrl)
			selector, err := NewSeqSelectorBuilder(providers).Build()
			assert.NoError(t, err)
			assert.NotNil(t, selector)

			var lastErr error
			for i := range tc.callTimes {
				p, selectErr := selector.Next(t.Context(), n)
				if selectErr != nil {
					lastErr = selectErr
				}

				switch {
				case i == tc.callTimes-1 && tc.wantErr != nil:
					assert.True(t, errors.Is(lastErr, tc.wantErr))
				case i < len(providers):
					assert.NoError(t, selectErr)
					assert.Equal(t, providers[i], p)
				default:
					assert.ErrorIs(t, selectErr, errs.ErrRecordNotFound)
					assert.Nil(t, p)
				}
			}
		})
	}
}
