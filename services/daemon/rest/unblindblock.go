// Copyright © 2024, 2025 Attestant Limited.
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
	"fmt"
	"io"
	"net/http"
	"strings"

	relay "github.com/attestantio/go-block-relay"
	"github.com/attestantio/go-eth2-client/api"
	apiv1bellatrix "github.com/attestantio/go-eth2-client/api/v1/bellatrix"
	apiv1capella "github.com/attestantio/go-eth2-client/api/v1/capella"
	apiv1deneb "github.com/attestantio/go-eth2-client/api/v1/deneb"
	apiv1electra "github.com/attestantio/go-eth2-client/api/v1/electra"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/pkg/errors"
)

func (s *Service) postUnblindBlock(w http.ResponseWriter, r *http.Request) {
	s.log.Trace().Msg("unblindBlock called")
	ctx := r.Context()

	signedBlindedBeaconBlock, err := s.obtainUnblindedBlock(ctx, r)
	if err != nil {
		s.log.Error().Err(err).Msg("Unable to obtain unblinded block")
		s.sendResponse(w, http.StatusInternalServerError, &APIResponse{
			Code:    http.StatusBadRequest,
			Message: "Unable to obtain blinded block",
		})
		monitorRequestHandled("unblind block", "failure")

		return
	}

	signedProposal, err := s.blockUnblinder.UnblindBlock(r.Context(), signedBlindedBeaconBlock)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, relay.ErrInvalidOptions) {
			code = http.StatusBadRequest
		}
		s.log.Error().Err(err).Msg("Failed to unblind block")
		s.sendResponse(w, http.StatusInternalServerError, &APIResponse{
			Code:    code,
			Message: "Failed to unblind block",
		})
		monitorRequestHandled("unblind block", "failure")

		return
	}

	if signedProposal == nil {
		s.sendResponse(w, http.StatusNoContent, nil)
		monitorRequestHandled("unblind block", "success")

		return
	}

	data, err := s.outputUnblindedBlock(ctx, signedProposal)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate output")
		s.sendResponse(w, http.StatusInternalServerError, &APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to unblind block",
		})
		monitorRequestHandled("unblind block", "failure")

		return
	}

	monitorRequestHandled("unblind block", "success")
	s.sendResponse(w, http.StatusOK, data)
}

func (s *Service) obtainUnblindedBlock(ctx context.Context,
	r *http.Request,
) (
	*api.VersionedSignedBlindedBeaconBlock,
	error,
) {
	contentType := s.obtainContentType(ctx, r)

	for k, v := range r.Header {
		s.log.Trace().Str("key", k).Strs("values", v).Msg("Header")
	}

	// Obtain the consensus version so we know what we have to unmarshal to.
	consensusVersions, exists := r.Header["Eth-Consensus-Version"]
	var consensusVersion string
	if !exists || len(consensusVersions) == 0 {
		s.log.Error().Msg("No Eth-Consensus-Version header")

		return nil, errors.New("No Eth-Consensus-Version header provided")
	}
	consensusVersion = consensusVersions[0]

	signedBlindedBeaconBlock, err := s.unmarshalBlindedBlock(ctx, contentType, consensusVersion, r)
	if err != nil {
		return nil, err
	}

	return signedBlindedBeaconBlock, nil
}

func (s *Service) unmarshalBlindedBlock(ctx context.Context,
	contentType string,
	consensusVersion string,
	r *http.Request,
) (
	*api.VersionedSignedBlindedBeaconBlock,
	error,
) {
	signedBlindedBeaconBlock := &api.VersionedSignedBlindedBeaconBlock{}

	switch strings.ToLower(consensusVersion) {
	case "bellatrix":
		signedBlindedBeaconBlock.Version = spec.DataVersionBellatrix
		signedBlindedBeaconBlock.Bellatrix = &apiv1bellatrix.SignedBlindedBeaconBlock{}
	case "capella":
		signedBlindedBeaconBlock.Version = spec.DataVersionCapella
		signedBlindedBeaconBlock.Capella = &apiv1capella.SignedBlindedBeaconBlock{}
	case "deneb":
		signedBlindedBeaconBlock.Version = spec.DataVersionDeneb
		signedBlindedBeaconBlock.Deneb = &apiv1deneb.SignedBlindedBeaconBlock{}
	case "electra":
		signedBlindedBeaconBlock.Version = spec.DataVersionElectra
		signedBlindedBeaconBlock.Electra = &apiv1electra.SignedBlindedBeaconBlock{}
	default:
		return nil, fmt.Errorf("unknown block version %v", consensusVersion)
	}

	switch strings.ToLower(contentType) {
	case "application/octet-stream":
		return s.unmarshalBlindedBlockSSZ(ctx, signedBlindedBeaconBlock, r.Body)
	case "application/json":
		return s.unmarshalBlindedBlockJSON(ctx, signedBlindedBeaconBlock, r.Body)
	default:
		return nil, fmt.Errorf("unsupported content type %s", contentType)
	}
}

func (s *Service) unmarshalBlindedBlockJSON(_ context.Context,
	signedBlindedBeaconBlock *api.VersionedSignedBlindedBeaconBlock,
	body io.Reader,
) (
	*api.VersionedSignedBlindedBeaconBlock,
	error,
) {
	var err error

	switch signedBlindedBeaconBlock.Version {
	case spec.DataVersionBellatrix:
		err = json.NewDecoder(body).Decode(signedBlindedBeaconBlock.Bellatrix)
	case spec.DataVersionCapella:
		err = json.NewDecoder(body).Decode(signedBlindedBeaconBlock.Capella)
	case spec.DataVersionDeneb:
		err = json.NewDecoder(body).Decode(signedBlindedBeaconBlock.Deneb)
	case spec.DataVersionElectra:
		err = json.NewDecoder(body).Decode(signedBlindedBeaconBlock.Electra)
	default:
		err = fmt.Errorf("unsupported block version %v", signedBlindedBeaconBlock.Version)
	}

	if err != nil {
		return nil, err
	}

	return signedBlindedBeaconBlock, nil
}

func (s *Service) unmarshalBlindedBlockSSZ(_ context.Context,
	signedBlindedBeaconBlock *api.VersionedSignedBlindedBeaconBlock,
	body io.Reader,
) (
	*api.VersionedSignedBlindedBeaconBlock,
	error,
) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}

	switch signedBlindedBeaconBlock.Version {
	case spec.DataVersionBellatrix:
		err = signedBlindedBeaconBlock.Bellatrix.UnmarshalSSZ(data)
	case spec.DataVersionCapella:
		err = signedBlindedBeaconBlock.Capella.UnmarshalSSZ(data)
	case spec.DataVersionDeneb:
		err = signedBlindedBeaconBlock.Deneb.UnmarshalSSZ(data)
	case spec.DataVersionElectra:
		err = signedBlindedBeaconBlock.Electra.UnmarshalSSZ(data)
	default:
		err = fmt.Errorf("unsupported block version %v", signedBlindedBeaconBlock.Version)
	}

	if err != nil {
		return nil, err
	}

	return signedBlindedBeaconBlock, nil
}

type unblindBlockResponse struct {
	Version spec.DataVersion          `json:"version"`
	Data    *unblindBlockResponseData `json:"data"`
}

type unblindBlockResponseData struct {
	ExecutionPayload *deneb.ExecutionPayload          `json:"execution_payload"`
	BlobsBundle      *unblindBlockResponseBlobsBundle `json:"blobs_bundle"`
}

type unblindBlockResponseBlobsBundle struct {
	Commitments []deneb.KZGCommitment `json:"commitments"`
	Proofs      []deneb.KZGProof      `json:"proofs"`
	Blobs       []deneb.Blob          `json:"blobs"`
}

func (s *Service) outputUnblindedBlock(_ context.Context,
	proposal *api.VersionedSignedProposal,
) (
	*unblindBlockResponse,
	error,
) {
	resp := &unblindBlockResponse{}
	resp.Version = proposal.Version
	switch resp.Version {
	case spec.DataVersionDeneb:
		resp.Data = &unblindBlockResponseData{
			ExecutionPayload: proposal.Deneb.SignedBlock.Message.Body.ExecutionPayload,
			BlobsBundle: &unblindBlockResponseBlobsBundle{
				Commitments: proposal.Deneb.SignedBlock.Message.Body.BlobKZGCommitments,
				Proofs:      proposal.Deneb.KZGProofs,
				Blobs:       proposal.Deneb.Blobs,
			},
		}
	case spec.DataVersionElectra:
		resp.Data = &unblindBlockResponseData{
			ExecutionPayload: proposal.Electra.SignedBlock.Message.Body.ExecutionPayload,
			BlobsBundle: &unblindBlockResponseBlobsBundle{
				Commitments: proposal.Electra.SignedBlock.Message.Body.BlobKZGCommitments,
				Proofs:      proposal.Electra.KZGProofs,
				Blobs:       proposal.Electra.Blobs,
			},
		}
	default:
		return nil, fmt.Errorf("unsupported version %v", resp.Version)
	}

	return resp, nil
}
