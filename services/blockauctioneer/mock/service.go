// Copyright © 2022, 2023 Attestant Limited.
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

package mock

import (
	"context"
	"math/big"

	"github.com/attestantio/go-block-relay/services/blockauctioneer"
	builderclient "github.com/attestantio/go-builder-client"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// Service is a mock block auctioneer.
type Service struct{}

// New creates a new mock block auctioneer.
func New() *Service {
	return &Service{}
}

// AuctionBlock obtains the best available use of the block space.
func (s *Service) AuctionBlock(_ context.Context,
	_ phase0.Slot,
	_ phase0.Hash32,
	_ phase0.BLSPubKey,
) (
	*blockauctioneer.Results,
	error,
) {
	return &blockauctioneer.Results{
		Values:    make(map[string]*big.Int),
		Providers: make([]builderclient.BuilderBidProvider, 0),
	}, nil
}
