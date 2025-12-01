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

	"github.com/attestantio/go-block-relay/services/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var metricsNamespace = "blockrelay"

var requests *prometheus.CounterVec

func registerMetrics(ctx context.Context, monitor metrics.Service) error {
	if requests != nil {
		// Already registered.
		return nil
	}

	if monitor == nil {
		// No monitor.
		return nil
	}

	if monitor.Presenter() == "prometheus" {
		return registerPrometheusMetrics(ctx)
	}

	return nil
}

func registerPrometheusMetrics(_ context.Context) error {
	requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "requests_total",
		Help:      "Requests",
	}, []string{"request", "result"})

	err := prometheus.Register(requests)
	if err != nil {
		return errors.Wrap(err, "failed to register requests_total")
	}

	return nil
}

func monitorRequestHandled(request string, result string) {
	if requests != nil {
		requests.WithLabelValues(request, result).Inc()
	}
}
