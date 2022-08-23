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
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gorilla/mux"
)

func (s *Service) getBuilderBid(w http.ResponseWriter, r *http.Request) {
	log.Trace().Msg("getBuilderBid called")
	ctx := context.Background()

	// Obtain path variables.
	vars := mux.Vars(r)
	tmpInt, err := strconv.ParseUint(vars["slot"], 10, 64)
	if err != nil {
		log.Trace().Err(err).Str("slot", vars["slot"]).Msg("Invalid slot")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slot := phase0.Slot(tmpInt)
	tmpBytes, err := hex.DecodeString(strings.TrimPrefix(vars["parenthash"], "0x"))
	if err != nil {
		log.Trace().Err(err).Str("parenthash", vars["parenthash"]).Msg("Invalid parent hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	parentHash := phase0.Hash32{}
	copy(parentHash[:], tmpBytes)
	tmpBytes, err = hex.DecodeString(strings.TrimPrefix(vars["pubkey"], "0x"))
	if err != nil {
		log.Trace().Err(err).Str("pubkey", vars["pubkey"]).Msg("Invalid public key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pubkey := phase0.BLSPubKey{}
	copy(pubkey[:], tmpBytes)

	bid, err := s.builderBidProvider.BuilderBid(ctx, slot, parentHash, pubkey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain bid")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(bid)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal bid")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Error().Err(err).Msg("Failed to write bid")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
