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
	"encoding/json"
	"net/http"

	"github.com/attestantio/go-block-relay/services/validatorregistrar"
	"github.com/attestantio/go-block-relay/types"
)

func (s *Service) postValidatorRegistrations(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if provider, isProvider := s.validatorRegistrar.(validatorregistrar.ValidatorRegistrationPassthrough); isProvider {
		// We have a passthrough: use it.
		if err := provider.ValidatorRegistrationsPassthrough(ctx, r.Body); err != nil {
			log.Error().Err(err).Msg("Failed to register validators with passthrough")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if provider, isProvider := s.validatorRegistrar.(validatorregistrar.ValidatorRegistrationHandler); isProvider {
		// We need to unmarshal the request body ourselves.

		registrations := make([]*types.SignedValidatorRegistration, 0)
		if err := json.NewDecoder(r.Body).Decode(&registrations); err != nil {
			log.Debug().Err(err).Msg("Supplied with invalid data")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := provider.ValidatorRegistrations(ctx, registrations); err != nil {
			log.Error().Err(err).Msg("Failed to register validators")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Error().Msg("Request not supported by service")
	w.WriteHeader(http.StatusInternalServerError)
}
