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
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/attestantio/go-block-relay/loggers"
	"github.com/attestantio/go-block-relay/services/blockunblinder"
	"github.com/attestantio/go-block-relay/services/builderbidprovider"
	"github.com/attestantio/go-block-relay/services/validatorregistrar"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
)

// Service is the REST daemon service.
type Service struct {
	log                zerolog.Logger
	srv                *http.Server
	validatorRegistrar validatorregistrar.Service
	builderBidProvider builderbidprovider.Service
	blockUnblinder     blockunblinder.Service
}

// New creates a new REST daemon service.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log := zerologger.With().Str("service", "daemon").Str("impl", "rest").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	if err := registerMetrics(ctx, parameters.monitor); err != nil {
		return nil, errors.New("failed to register metrics")
	}

	s := &Service{
		log:                log,
		validatorRegistrar: parameters.validatorRegistrar,
		builderBidProvider: parameters.builderBidProvider,
		blockUnblinder:     parameters.blockUnblinder,
	}

	if err := s.startServer(ctx, parameters.serverName, parameters.listenAddress); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) startServer(ctx context.Context,
	_ string,
	listenAddress string,
) error {
	// Set to release mode to remove debug logging.
	gin.SetMode(gin.ReleaseMode)

	// Start up the router.
	r := gin.New()
	r.Use(gin.Recovery())
	if err := r.SetTrustedProxies(nil); err != nil {
		return errors.Wrap(err, "failed to set trusted proxies")
	}
	r.Use(loggers.NewGinLogger(s.log))

	router := mux.NewRouter()
	router.HandleFunc("/eth/v1/builder/validators", s.postValidatorRegistrations).Methods("POST")
	router.HandleFunc("/eth/v1/builder/header/{slot}/{parenthash}/{pubkey}", s.getBuilderBid).Methods("GET")
	router.HandleFunc("/eth/v1/builder/status", s.getStatus).Methods("GET")
	router.HandleFunc("/eth/v1/builder/blinded_blocks", s.postUnblindBlock).Methods("POST")
	router.PathPrefix("/").Handler(s)

	s.srv = &http.Server{
		Addr:              listenAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	// At current the service does not run over HTTPS.
	//	if false {
	//		certManager := autocert.Manager{
	//			Prompt:     autocert.AcceptTOS,
	//			HostPolicy: autocert.HostWhitelist(serverName),
	//			Cache:      autocert.DirCache("./certs"),
	//		}
	//
	//		s.srv.TLSConfig = &tls.Config{
	//			MinVersion:               tls.VersionTLS13,
	//			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
	//			GetCertificate:           certManager.GetCertificate,
	//			PreferServerCipherSuites: true,
	//			CipherSuites: []uint16{
	//				tls.TLS_AES_128_GCM_SHA256,
	//				tls.TLS_CHACHA20_POLY1305_SHA256,
	//				tls.TLS_AES_256_GCM_SHA384,
	//				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	//				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	//				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	//				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	//				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	//				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	//			},
	//		}
	//		s.srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	//
	//		// Listen on HTTP port for certificate updates.
	//		go func() {
	//			s.log.Trace().Str("listen_address", listenAddress).Msg("Starting certificate update service")
	//			server := &http.Server{
	//				Addr:              ":http",
	//				Handler:           certManager.HTTPHandler(nil),
	//				ReadHeaderTimeout: 5 * time.Second,
	//			}
	//			if err := server.ListenAndServe(); err != nil {
	//				s.log.Error().Err(err).Msg("Certificate update service stopped")
	//			}
	//		}()
	//
	//		go func() {
	//			s.log.Trace().Str("listen_address", listenAddress).Msg("Starting HTTPS daemon")
	//			if err := s.srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
	//				// if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
	//				s.log.Error().Err(err).Msg("Server shut down unexpectedly")
	//			}
	//		}()
	//	} else {
	// Insecure.
	go func() {
		s.log.Trace().Str("listen_address", listenAddress).Msg("Starting HTTP daemon")
		if err := s.srv.ListenAndServe(); err != nil {
			s.log.Error().Err(err).Msg("HTTP server shut down")
		}
	}()
	// }

	go s.sigloop(ctx)

	return nil
}

func (s *Service) sigloop(ctx context.Context) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	for {
		select {
		case sig := <-sigCh:
			if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == os.Interrupt || sig == os.Kill {
				s.log.Info().Msg("Received signal, shutting down")
				if err := s.srv.Shutdown(ctx); err != nil {
					s.log.Warn().Err(err).Msg("Failed to shutdown service")
				}

				return
			}
		case <-ctx.Done():
			s.log.Info().Msg("Context done, shutting down")
			if err := s.srv.Shutdown(ctx); err != nil {
				s.log.Warn().Err(err).Msg("Failed to shutdown service")
			}

			return
		}
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log.Debug().Str("method", r.Method).Stringer("url", r.URL).Msg("Unhandled request")

	w.WriteHeader(http.StatusNotFound)
}

func (s *Service) obtainContentType(_ context.Context,
	r *http.Request,
) string {
	contentTypeHeaderVals, exists := r.Header["Content-Type"]
	var contentType string
	if !exists {
		// Assume that no content type == JSON, for backwards-compatibility.
		contentType = "application/json"
	} else {
		contentType = contentTypeHeaderVals[0]
	}

	// Remove supplementary information.
	index := strings.Index(contentType, ";")
	if index > 0 {
		contentType = contentType[:index]
	}
	contentType = strings.TrimSpace(contentType)

	return contentType
}
