// Copyright © 2022 Attestant Limited.
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

package rest_test

import (
	"context"
	"testing"

	mockauctioneer "github.com/attestantio/go-block-relay/services/blockauctioneer/mock"
	restdaemon "github.com/attestantio/go-block-relay/services/daemon/rest"
	nullmetrics "github.com/attestantio/go-block-relay/services/metrics/null"
	mockregistrar "github.com/attestantio/go-block-relay/services/validatorregistrar/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	ctx := context.Background()
	registrar := mockregistrar.New()
	auctioneer := mockauctioneer.New()
	monitor := nullmetrics.New()

	tests := []struct {
		name   string
		params []restdaemon.Parameter
		err    string
	}{
		{
			name: "MonitorMissing",
			params: []restdaemon.Parameter{
				restdaemon.WithLogLevel(zerolog.Disabled),
				restdaemon.WithMonitor(nil),
				restdaemon.WithServerName("server.attestant.io"),
				restdaemon.WithListenAddress(":14734"),
				restdaemon.WithValidatorRegistrar(registrar),
				restdaemon.WithBlockAuctioneer(auctioneer),
			},
			err: "problem with parameters: no monitor specified",
		},
		{
			name: "ServerNameMissing",
			params: []restdaemon.Parameter{
				restdaemon.WithLogLevel(zerolog.Disabled),
				restdaemon.WithMonitor(monitor),
				restdaemon.WithListenAddress(":14734"),
				restdaemon.WithValidatorRegistrar(registrar),
				restdaemon.WithBlockAuctioneer(auctioneer),
			},
			err: "problem with parameters: no server name specified",
		},
		{
			name: "ListenAddressMissing",
			params: []restdaemon.Parameter{
				restdaemon.WithLogLevel(zerolog.Disabled),
				restdaemon.WithMonitor(monitor),
				restdaemon.WithServerName("server.attestant.io"),
				restdaemon.WithValidatorRegistrar(registrar),
				restdaemon.WithBlockAuctioneer(auctioneer),
			},
			err: "problem with parameters: no listen address specified",
		},
		{
			name: "RegistrarMissing",
			params: []restdaemon.Parameter{
				restdaemon.WithLogLevel(zerolog.Disabled),
				restdaemon.WithMonitor(monitor),
				restdaemon.WithServerName("server.attestant.io"),
				restdaemon.WithListenAddress(":14734"),
				restdaemon.WithBlockAuctioneer(auctioneer),
			},
			err: "problem with parameters: no validator registrar specified",
		},
		{
			name: "AuctioneerMissing",
			params: []restdaemon.Parameter{
				restdaemon.WithLogLevel(zerolog.Disabled),
				restdaemon.WithMonitor(monitor),
				restdaemon.WithServerName("server.attestant.io"),
				restdaemon.WithListenAddress(":14734"),
				restdaemon.WithValidatorRegistrar(registrar),
			},
			err: "problem with parameters: no block auctioneer specified",
		},
		{
			name: "Good",
			params: []restdaemon.Parameter{
				restdaemon.WithLogLevel(zerolog.Disabled),
				restdaemon.WithMonitor(monitor),
				restdaemon.WithServerName("server.attestant.io"),
				restdaemon.WithListenAddress(":14734"),
				restdaemon.WithValidatorRegistrar(registrar),
				restdaemon.WithBlockAuctioneer(auctioneer),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := restdaemon.New(ctx, test.params...)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
