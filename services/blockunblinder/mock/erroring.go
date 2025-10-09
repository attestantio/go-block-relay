// Copyright © 2024 Attestant Limited.
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

// Package mock provides mock implementations for testing.
package mock

import (
	"context"
	"errors"

	"github.com/attestantio/go-eth2-client/api"
)

// ErroringService is a mock block unblinder.
type ErroringService struct{}

// NewErroring creates a new mock block unblinder.
func NewErroring() *ErroringService {
	return &ErroringService{}
}

// UnblindBlock unblinds the given block.
func (s *ErroringService) UnblindBlock(_ context.Context,
	_ *api.VersionedSignedBlindedBeaconBlock,
) (
	*api.VersionedSignedProposal,
	error,
) {
	return nil, errors.New("error")
}
