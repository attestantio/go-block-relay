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

package blockauctioneer

import (
	"context"
	"math/big"

	builderclient "github.com/attestantio/go-builder-client"
	"github.com/attestantio/go-builder-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// Service defines the block auctioneer service.
type Service interface{}

// Results provides the results of the auction process.
type Results struct {
	// Values is a map of builder client=>value.
	Values map[string]*big.Int
	// Providers is used to unblind the block.
	Providers []builderclient.BuilderBidProvider
	// Bid is the winning signed builder bid.
	Bid *spec.VersionedSignedBuilderBid
}

// BlockAuctioneer is the interface for auctioning block space.
type BlockAuctioneer interface {
	// AuctionBlock obtains the best available use of the block space.
	AuctionBlock(ctx context.Context,
		slot phase0.Slot,
		parentHash phase0.Hash32,
		pubkey phase0.BLSPubKey,
	) (
		*Results,
		error,
	)
}
