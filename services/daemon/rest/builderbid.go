// Copyright © 2022, 204 Attestant Limited.
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
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gorilla/mux"
)

func (s *Service) getBuilderBid(w http.ResponseWriter, r *http.Request) {
	s.log.Trace().Msg("getBuilderBid called")

	// Obtain path variables.
	vars := mux.Vars(r)
	tmpInt, err := strconv.ParseUint(vars["slot"], 10, 64)
	if err != nil {
		s.log.Debug().Err(err).Str("slot", vars["slot"]).Msg("Invalid slot")
		s.sendResponse(w, http.StatusBadRequest, &APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("invalid slot %s", vars["slot"]),
		})
		monitorRequestHandled("builder bid", "failure")

		return
	}
	slot := phase0.Slot(tmpInt)
	tmpBytes, err := hex.DecodeString(strings.TrimPrefix(vars["parenthash"], "0x"))
	if err != nil {
		s.log.Debug().Err(err).Str("parenthash", vars["parenthash"]).Msg("Invalid parent hash")
		s.sendResponse(w, http.StatusBadRequest, &APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("invalid parent hash %s", vars["parenthash"]),
		})
		monitorRequestHandled("builder bid", "failure")

		return
	}
	parentHash := phase0.Hash32{}
	copy(parentHash[:], tmpBytes)
	tmpBytes, err = hex.DecodeString(strings.TrimPrefix(vars["pubkey"], "0x"))
	if err != nil {
		s.log.Trace().Err(err).Str("pubkey", vars["pubkey"]).Msg("Invalid public key")
		s.sendResponse(w, http.StatusBadRequest, &APIResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("invalid public key %s", vars["pubkey"]),
		})
		monitorRequestHandled("builder bid", "failure")

		return
	}
	pubkey := phase0.BLSPubKey{}
	copy(pubkey[:], tmpBytes)

	bid, err := s.builderBidProvider.BuilderBid(r.Context(), slot, parentHash, pubkey)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to obtain bid")
		s.sendResponse(w, http.StatusInternalServerError, &APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to obtain bid",
		})
		monitorRequestHandled("builder bid", "failure")

		return
	}

	monitorRequestHandled("builder bid", "success")

	if bid == nil {
		s.sendResponse(w, http.StatusNoContent, nil)
	} else {
		s.sendResponse(w, http.StatusOK, bid)
	}
}
