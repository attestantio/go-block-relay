// Copyright © 2025 Attestant Limited.
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
	"context"
	"errors"
	"net/http"

	relay "github.com/attestantio/go-block-relay"
	"github.com/attestantio/go-eth2-client/api"
)

func (s *Service) submitBlindedBlock(w http.ResponseWriter, r *http.Request) {
	s.log.Trace().Msg("submitBlindedBlock called")
	ctx := r.Context()

	signedBlindedBeaconBlock, err := s.obtainSignedBlindedBlock(ctx, r)
	if err != nil {
		s.log.Error().Err(err).Msg("Unable to obtain signed blinded block")
		s.sendResponse(w,
			http.StatusInternalServerError,
			map[string]string{},
			&APIResponse{
				Code:    http.StatusBadRequest,
				Message: "Unable to obtain signed blinded block",
			})
		monitorRequestHandled("submit blinded block", "failure")

		return
	}

	err = s.blockSubmitter.SubmitBlock(r.Context(), signedBlindedBeaconBlock)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, relay.ErrInvalidOptions) {
			code = http.StatusBadRequest
		}
		s.log.Error().Err(err).Msg("Failed to submit blinded block")
		s.sendResponse(w,
			http.StatusInternalServerError,
			map[string]string{},
			&APIResponse{
				Code:    code,
				Message: "Failed to submit blinded block",
			})
		monitorRequestHandled("submit blinded block", "failure")

		return
	}

	s.sendResponse(w,
		http.StatusOK,
		map[string]string{},
		nil,
	)
}

func (s *Service) obtainSignedBlindedBlock(ctx context.Context,
	r *http.Request,
) (*api.VersionedSignedBlindedBeaconBlock,
	error,
) {
	// Reusing the code for obtaining the signed blinded block for blinded_blocks v1
	// but with updated semantics for clearer documentation.
	return s.obtainUnblindedBlock(ctx, r)
}
