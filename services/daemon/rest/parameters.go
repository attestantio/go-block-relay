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

package rest

import (
	"errors"

	"github.com/attestantio/go-block-relay/services/blockauctioneer"
	"github.com/attestantio/go-block-relay/services/builderbidprovider"
	"github.com/attestantio/go-block-relay/services/metrics"
	nullmetrics "github.com/attestantio/go-block-relay/services/metrics/null"
	"github.com/attestantio/go-block-relay/services/validatorregistrar"
	"github.com/rs/zerolog"
)

type parameters struct {
	logLevel           zerolog.Level
	monitor            metrics.Service
	serverName         string
	listenAddress      string
	validatorRegistrar validatorregistrar.Service
	blockAuctioneer    blockauctioneer.Service
	builderBidProvider builderbidprovider.Service
}

// Parameter is the interface for service parameters.
type Parameter interface {
	apply(p *parameters)
}

type parameterFunc func(*parameters)

func (f parameterFunc) apply(p *parameters) {
	f(p)
}

// WithLogLevel sets the log level for the module.
func WithLogLevel(logLevel zerolog.Level) Parameter {
	return parameterFunc(func(p *parameters) {
		p.logLevel = logLevel
	})
}

// WithMonitor sets the monitor for the module.
func WithMonitor(monitor metrics.Service) Parameter {
	return parameterFunc(func(p *parameters) {
		p.monitor = monitor
	})
}

// WithServerName sets the server name for this module.
func WithServerName(name string) Parameter {
	return parameterFunc(func(p *parameters) {
		p.serverName = name
	})
}

// WithListenAddress sets the listen address for this module.
func WithListenAddress(listenAddress string) Parameter {
	return parameterFunc(func(p *parameters) {
		p.listenAddress = listenAddress
	})
}

// WithValidatorRegistrar sets the validator registrar.
func WithValidatorRegistrar(validatorRegistrar validatorregistrar.Service) Parameter {
	return parameterFunc(func(p *parameters) {
		p.validatorRegistrar = validatorRegistrar
	})
}

// WithBuilderBidProvider sets the builder bid provider.
func WithBuilderBidProvider(provider builderbidprovider.Service) Parameter {
	return parameterFunc(func(p *parameters) {
		p.builderBidProvider = provider
	})
}

// WithBlockAuctioneer sets the block auctioneer.
func WithBlockAuctioneer(blockAuctioneer blockauctioneer.Service) Parameter {
	return parameterFunc(func(p *parameters) {
		p.blockAuctioneer = blockAuctioneer
	})
}

// parseAndCheckParameters parses and checks parameters to ensure that mandatory parameters are present and correct.
func parseAndCheckParameters(params ...Parameter) (*parameters, error) {
	parameters := parameters{
		logLevel: zerolog.GlobalLevel(),
		monitor:  nullmetrics.New(),
	}
	for _, p := range params {
		if params != nil {
			p.apply(&parameters)
		}
	}

	if parameters.monitor == nil {
		return nil, errors.New("no monitor specified")
	}
	// At current the server name is not required, as the daemon does not run on HTTPS.
	// if parameters.serverName == "" {
	// 	return nil, errors.New("no server name specified")
	// }
	if parameters.listenAddress == "" {
		return nil, errors.New("no listen address specified")
	}
	if parameters.validatorRegistrar == nil {
		return nil, errors.New("no validator registrar specified")
	}
	if parameters.blockAuctioneer == nil {
		return nil, errors.New("no block auctioneer specified")
	}
	if parameters.builderBidProvider == nil {
		return nil, errors.New("no builder bid provider specified")
	}

	return &parameters, nil
}
