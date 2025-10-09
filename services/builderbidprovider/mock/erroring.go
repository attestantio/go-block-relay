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

// Package mock provides mock implementations for testing.
package mock

import (
	"context"
	"errors"

	"github.com/attestantio/go-builder-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// ErroringService is a mock builder bid provider.
type ErroringService struct{}

// NewErroring creates a new mock builder bid provider.
func NewErroring() *ErroringService {
	return &ErroringService{}
}

// BuilderBid provides a builder bid.
func (s *ErroringService) BuilderBid(_ context.Context,
	_ phase0.Slot,
	_ phase0.Hash32,
	_ phase0.BLSPubKey,
) (
	*spec.VersionedSignedBuilderBid,
	error,
) {
	return nil, errors.New("error")
}
