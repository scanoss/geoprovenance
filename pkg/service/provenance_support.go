// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"encoding/json"
	"errors"

	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/provenancev2"
	"scanoss.com/provenance/pkg/dtos"

	"go.uber.org/zap"
)

// convertPurlRequestInput converts a Purl Request structure into an internal Provenance Input struct
func convertProvenanceInput(s *zap.SugaredLogger, request *common.PurlRequest) (dtos.ProvenanceInput, error) {
	data, err := json.Marshal(request)
	if err != nil {
		s.Errorf("Problem marshalling Provenance request input: %v", err)
		return dtos.ProvenanceInput{}, errors.New("problem marshalling Provenance input")
	}
	dtoRequest, err := dtos.ParseProvenanceInput(s, data)
	if err != nil {
		s.Errorf("Problem parsing Provenance request input: %v", err)
		return dtos.ProvenanceInput{}, errors.New("problem parsing Provenance input")
	}
	return dtoRequest, nil
}

// convertProvenanceOutput converts an internal Provenance Output structure into a Provenance Response struct
func convertProvenanceOutput(s *zap.SugaredLogger, output dtos.ProvenanceOutput) (*pb.ProvenanceResponse, error) {
	data, err := json.Marshal(output)
	if err != nil {
		s.Errorf("Problem marshalling Provenance request output: %v", err)
		return &pb.ProvenanceResponse{}, errors.New("problem marshalling Provenance output")
	}
	//zlog.S.Debugf("Parsed data: %v", string(data))
	var depResp pb.ProvenanceResponse
	err = json.Unmarshal(data, &depResp)
	if err != nil {
		s.Errorf("Problem unmarshalling Provenance request output: %v", err)
		return &pb.ProvenanceResponse{}, errors.New("problem unmarshalling Provenance output")
	}
	return &depResp, nil
}
