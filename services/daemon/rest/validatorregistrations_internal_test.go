// Copyright © 2022, 2025 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockauctioneer "github.com/attestantio/go-block-relay/services/blockauctioneer/mock"
	mockblockunblinder "github.com/attestantio/go-block-relay/services/blockunblinder/mock"
	mockbuilderbidprovider "github.com/attestantio/go-block-relay/services/builderbidprovider/mock"
	nullmetrics "github.com/attestantio/go-block-relay/services/metrics/null"
	mockvalidatorregistrar "github.com/attestantio/go-block-relay/services/validatorregistrar/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestValidatorRegistrations(t *testing.T) {
	ctx := context.Background()

	registrar := mockvalidatorregistrar.New()
	auctioneer := mockauctioneer.New()
	unblinder := mockblockunblinder.New()
	monitor := nullmetrics.New()
	builderBidProvider := mockbuilderbidprovider.New()

	service, err := New(ctx,
		WithLogLevel(zerolog.Disabled),
		WithMonitor(monitor),
		WithServerName("server.attestant.io"),
		WithListenAddress(":14734"),
		WithValidatorRegistrar(registrar),
		WithBlockAuctioneer(auctioneer),
		WithBlockUnblinder(unblinder),
		WithBuilderBidProvider(builderBidProvider),
	)
	require.NoError(t, err)

	erroringRegistrar := mockvalidatorregistrar.NewErroring()
	erroringService, err := New(ctx,
		WithLogLevel(zerolog.Disabled),
		WithMonitor(monitor),
		WithServerName("server.attestant.io"),
		WithListenAddress(":14735"),
		WithValidatorRegistrar(registrar),
		WithValidatorRegistrar(erroringRegistrar),
		WithBlockAuctioneer(auctioneer),
		WithBlockUnblinder(unblinder),
		WithBuilderBidProvider(builderBidProvider),
	)
	require.NoError(t, err)

	tests := []struct {
		name       string
		service    *Service
		request    *http.Request
		writer     *httptest.ResponseRecorder
		statusCode int
	}{
		{
			name:       "Good",
			service:    service,
			request:    &http.Request{},
			writer:     httptest.NewRecorder(),
			statusCode: http.StatusOK,
		},
		{
			name:       "Erroring",
			service:    erroringService,
			request:    &http.Request{},
			writer:     httptest.NewRecorder(),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.service.postValidatorRegistrations(test.writer, test.request)
			require.Equal(t, test.statusCode, test.writer.Result().StatusCode)
		})
	}
}

func TestValidatorRegistrationsContentType(t *testing.T) {
	ctx := context.Background()

	registrar := mockvalidatorregistrar.NewHandler()
	auctioneer := mockauctioneer.New()
	unblinder := mockblockunblinder.New()
	monitor := nullmetrics.New()
	builderBidProvider := mockbuilderbidprovider.New()

	service, err := New(ctx,
		WithLogLevel(zerolog.Disabled),
		WithMonitor(monitor),
		WithServerName("server.attestant.io"),
		WithListenAddress(":14734"),
		WithValidatorRegistrar(registrar),
		WithBlockAuctioneer(auctioneer),
		WithBlockUnblinder(unblinder),
		WithBuilderBidProvider(builderBidProvider),
	)
	require.NoError(t, err)

	tests := []struct {
		name       string
		service    *Service
		request    *http.Request
		writer     *httptest.ResponseRecorder
		statusCode int
	}{
		{
			name:    "NoContentType",
			service: service,
			request: &http.Request{
				Header: map[string][]string{},
				Body:   io.NopCloser(bytes.NewReader([]byte("[]"))),
			},
			writer:     httptest.NewRecorder(),
			statusCode: http.StatusOK,
		},
		{
			name:    "JSON",
			service: service,
			request: &http.Request{
				Header: map[string][]string{
					"Content": {"application/json"},
				},
				Body: io.NopCloser(bytes.NewReader([]byte("[]"))),
			},
			writer:     httptest.NewRecorder(),
			statusCode: http.StatusOK,
		},
		{
			name:    "JSONCharset",
			service: service,
			request: &http.Request{
				Header: map[string][]string{
					"Content-Type": {"application/json; charset=utf-8"},
				},
				Body: io.NopCloser(bytes.NewReader([]byte("[]"))),
			},
			writer:     httptest.NewRecorder(),
			statusCode: http.StatusOK,
		},
		{
			name:    "Bad",
			service: service,
			request: &http.Request{
				Header: map[string][]string{
					"Content-Type": {"application/garbage"},
				},
				Body: io.NopCloser(bytes.NewReader([]byte("[]"))),
			},
			writer:     httptest.NewRecorder(),
			statusCode: http.StatusUnsupportedMediaType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.service.postValidatorRegistrations(test.writer, test.request)
			require.Equal(t, test.statusCode, test.writer.Result().StatusCode)
		})
	}
}
