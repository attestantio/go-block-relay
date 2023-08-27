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
	"testing"

	nullmetrics "github.com/attestantio/go-block-relay/services/metrics/null"
	prometheusmetrics "github.com/attestantio/go-block-relay/services/metrics/prometheus"
	"github.com/stretchr/testify/require"
)

func TestRegisterMetrics(t *testing.T) {
	ctx := context.Background()

	// Ensure metrics handler can be called without failing.
	monitorRequestHandled("test", "success")

	// Ensure metrics can be registered without monitor.
	require.NoError(t, registerMetrics(ctx, nil))

	// Ensure metrics can be registered with a null monitor.
	nullMonitor := nullmetrics.New()
	require.NoError(t, registerMetrics(ctx, nullMonitor))

	// Ensure metrics can be registered with a prometheus monitor.
	monitor, err := prometheusmetrics.New(ctx,
		prometheusmetrics.WithAddress(":14632"),
	)
	require.NoError(t, err)
	require.NoError(t, registerMetrics(ctx, monitor))

	// Ensure metrics can be re-registered without error.
	require.NoError(t, registerMetrics(ctx, monitor))

	// Ensure intneral function recognises double registration and errors.
	require.EqualError(t, registerPrometheusMetrics(ctx), "failed to register requests_total: duplicate metrics collector registration attempted")

	// Ensure metrics handler can be called without failing.
	monitorRequestHandled("test", "success")
}
