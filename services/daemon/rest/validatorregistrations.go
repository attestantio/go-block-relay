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
	"net/http"

	"github.com/attestantio/go-block-relay/services/validatorregistrar"
)

func (s *Service) postValidatorRegistrations(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	provider, isProvider := s.validatorRegistrar.(validatorregistrar.ValidatorRegistrationPassthrough)
	if isProvider {
		// We have a passthrough: use it.
		if err := provider.ValidatorRegistrationsPassthrough(ctx, r.Body); err != nil {
			log.Error().Err(err).Msg("Failed to register validators with passthrough")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Error().Msg("Non-passthrough for validator registration not currently supported")
	w.WriteHeader(http.StatusInternalServerError)
}
