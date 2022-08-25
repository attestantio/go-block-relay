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
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/attestantio/go-block-relay/loggers"
	"github.com/attestantio/go-block-relay/services/builderbidprovider"
	"github.com/attestantio/go-block-relay/services/validatorregistrar"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme/autocert"
)

// Service is the REST daemon service.
type Service struct {
	srv                *http.Server
	validatorRegistrar validatorregistrar.Service
	builderBidProvider builderbidprovider.Service
}

// module-wide log.
var log zerolog.Logger

// New creates a new JSON-RPC daemon service.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log = zerologger.With().Str("service", "daemon").Str("impl", "rest").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	if err := registerMetrics(ctx, parameters.monitor); err != nil {
		return nil, errors.New("failed to register metrics")
	}

	s := &Service{
		validatorRegistrar: parameters.validatorRegistrar,
		builderBidProvider: parameters.builderBidProvider,
	}

	// Set to release mode to remove debug logging.
	gin.SetMode(gin.ReleaseMode)

	// Start up the router.
	r := gin.New()
	r.Use(gin.Recovery())
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, errors.Wrap(err, "failed to set trusted proxies")
	}
	r.Use(loggers.NewGinLogger(log))

	router := mux.NewRouter()
	router.HandleFunc("/eth/v1/builder/validators", s.postValidatorRegistrations).Methods("POST")
	router.HandleFunc("/eth/v1/builder/header/{slot}/{parenthash}/{pubkey}", s.getBuilderBid).Methods("GET")
	router.HandleFunc("/eth/v1/builder/status", s.getStatus).Methods("GET")

	// At current the service does not run over HTTPS.
	if false {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(parameters.serverName),
			Cache:      autocert.DirCache("./certs"),
		}

		s.srv = &http.Server{
			Addr:    parameters.listenAddress,
			Handler: router,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS13,
				CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
				GetCertificate:           certManager.GetCertificate,
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				},
			},
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}

		// Listen on HTTP port for certificate updates.
		go func() {
			log.Trace().Str("listen_address", parameters.listenAddress).Msg("Starting certificate update service")
			if err := http.ListenAndServe(":http", certManager.HTTPHandler(nil)); err != nil {
				log.Error().Err(err).Msg("Certificate update service stopped")
			}
		}()

		go func() {
			log.Trace().Str("listen_address", parameters.listenAddress).Msg("Starting daemon")
			if err := s.srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
				// if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Error().Err(err).Msg("Server shut down unexpectedly")
			}
		}()
	} else {
		// Insecure.
		go func() {
			log.Trace().Str("listen_address", parameters.listenAddress).Msg("Starting daemon")
			if err := http.ListenAndServe(parameters.listenAddress, router); err != nil {
				log.Error().Err(err).Msg("HTTP server shut down")
			}
		}()

	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		for {
			sig := <-sigCh
			if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == os.Interrupt || sig == os.Kill {
				if err := s.srv.Shutdown(ctx); err != nil {
					log.Warn().Err(err).Msg("Failed to shutdown service")
				}
				break
			}
		}
	}()

	return s, nil
}
