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

package validatorregistrar

import (
	"context"
	"io"

	"github.com/attestantio/go-block-relay/types"
)

// Service defines the validator registrar service.
type Service interface{}

// ValidatorRegistrationPassthrough is the interface for handling validator registrations with passthrough.
type ValidatorRegistrationPassthrough interface {
	// ValidatorRegistrationsPassthrough handles validator registrations directly.
	ValidatorRegistrationsPassthrough(ctx context.Context, reader io.ReadCloser) ([]string, error)
}

// ValidatorRegistrationHandler is the interface for handling validator registrations.
type ValidatorRegistrationHandler interface {
	// ValidatorRegistrations handles validator registrations.
	ValidatorRegistrations(ctx context.Context, registrations []*types.SignedValidatorRegistration) ([]string, error)
}
