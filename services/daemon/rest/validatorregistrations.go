// Copyright © 2022, 2024 Attestant Limited.
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
	"strings"

	relay "github.com/attestantio/go-block-relay"
	"github.com/attestantio/go-block-relay/services/validatorregistrar"
	"github.com/attestantio/go-block-relay/types"
	"github.com/pkg/errors"
)

func (s *Service) postValidatorRegistrations(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	var registrationErrors []string
	var err error

	passthroughProvider, isPassthroughProvider := s.validatorRegistrar.(validatorregistrar.ValidatorRegistrationPassthrough)
	handler, isHandler := s.validatorRegistrar.(validatorregistrar.ValidatorRegistrationHandler)
	switch {
	case isPassthroughProvider:
		statusCode, registrationErrors, err = s.postValidatorRegistrationsPassthrough(r.Context(), r, passthroughProvider)
	case isHandler:
		statusCode, registrationErrors, err = s.postValidatorRegistrationsHandler(r.Context(), r, handler)
	default:
		s.log.Error().Msg("Request not supported by service")
		err = errors.New("Request not supported by service")
	}

	if err != nil {
		s.sendResponse(w, http.StatusBadRequest, &APIResponse{
			Code:    statusCode,
			Message: err.Error(),
		})
		monitorRequestHandled("validator registrations", "failure")

		return
	}

	monitorRequestHandled("validator registrations", "success")
	if len(registrationErrors) == 0 {
		s.sendResponse(w, http.StatusOK, nil)
	} else {
		s.sendResponse(w, http.StatusBadRequest, &APIResponse{
			Code:    http.StatusBadRequest,
			Message: strings.Join(registrationErrors, ";"),
		})
	}
}

func (s *Service) postValidatorRegistrationsPassthrough(ctx context.Context,
	r *http.Request,
	provider validatorregistrar.ValidatorRegistrationPassthrough,
) (
	int,
	[]string,
	error,
) {
	registrationErrors, err := provider.ValidatorRegistrationsPassthrough(ctx, r.Body)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to register validators with passthrough")
		code := http.StatusInternalServerError
		if errors.Is(err, relay.ErrInvalidOptions) {
			code = http.StatusBadRequest
		}

		return code, nil, errors.New("failed to register validators")
	}

	return http.StatusOK, registrationErrors, nil
}

func (s *Service) postValidatorRegistrationsHandler(ctx context.Context,
	r *http.Request,
	provider validatorregistrar.ValidatorRegistrationHandler,
) (
	int,
	[]string,
	error,
) {
	// We need to unmarshal the request body ourselves.
	registrations := make([]*types.SignedValidatorRegistration, 0)
	if err := json.NewDecoder(r.Body).Decode(&registrations); err != nil {
		s.log.Debug().Err(err).Msg("Supplied with invalid data")

		return http.StatusBadRequest, nil, errors.Wrap(err, "invalid JSON")
	}

	registrationErrors, err := provider.ValidatorRegistrations(ctx, registrations)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to register validators")
		code := http.StatusInternalServerError
		if errors.Is(err, relay.ErrInvalidOptions) {
			code = http.StatusBadRequest
		}

		return code, nil, errors.Wrap(err, "failed to register validators")
	}

	return http.StatusOK, registrationErrors, nil
}
