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

	"github.com/scanoss/go-component-helper/componenthelper"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/geoprovenancev2"
	"go.uber.org/zap"
	"scanoss.com/provenance/pkg/dtos"
	se "scanoss.com/provenance/pkg/errors"
)

// convertPurlRequestInput converts a Purl Request structure into an internal Provenance Input struct.
func convertProvenanceInput(request *common.PurlRequest) ([]componenthelper.ComponentDTO, error) {
	if request == nil || len(request.Purls) == 0 {
		return []componenthelper.ComponentDTO{}, se.NewBadRequestError("Empty purls provided", nil)
	}
	var componentDTOS []componenthelper.ComponentDTO
	emptyPurlsCount := 0
	for _, c := range request.Purls {
		if c.Purl == "" {
			emptyPurlsCount++
			continue
		}
		componentDTOS = append(componentDTOS, componenthelper.ComponentDTO{
			Purl:        c.Purl,
			Requirement: c.Requirement,
		})
	}
	if emptyPurlsCount > 0 && emptyPurlsCount == len(request.Purls) {
		return []componenthelper.ComponentDTO{}, se.NewBadRequestError("Empty purls provided", nil)
	}
	return componentDTOS, nil
}

// convertProvenanceOutput converts an internal Provenance Output structure into a Provenance Response struct.
func convertProvenanceOutput(s *zap.SugaredLogger, output dtos.ProvenanceOutput) (*pb.ContributorResponse, error) {
	response := pb.ContributorResponse{}
	for _, p := range output.Provenance {
		curatedData, err := json.Marshal(p.CuratedLocations)
		if err != nil {
			s.Errorf("Problem marshalling curated locations for %s: %v", p.Purl, err)
			return &pb.ContributorResponse{}, errors.New("problem marshalling curated locations")
		}
		var curatedLocations []*pb.CuratedLocation
		err = json.Unmarshal(curatedData, &curatedLocations)
		if err != nil {
			s.Errorf("Problem unmarshalling curated locations for %s: %v", p.Purl, err)
			return &pb.ContributorResponse{}, errors.New("problem unmarshalling curated locations")
		}

		declaredData, err := json.Marshal(p.DeclaredLocations)
		if err != nil {
			s.Errorf("Problem marshalling declared locations for %s: %v", p.Purl, err)
			return &pb.ContributorResponse{}, errors.New("problem marshalling declared locations")
		}
		var declaredLocations []*pb.DeclaredLocation
		err = json.Unmarshal(declaredData, &declaredLocations)
		if err != nil {
			s.Errorf("Problem unmarshalling declared locations for %s: %v", p.Purl, err)
			return &pb.ContributorResponse{}, errors.New("problem unmarshalling declared locations")
		}
		contributorsResponse := &pb.ContributorResponse_Purls{
			Purl:              p.Purl,
			DeclaredLocations: declaredLocations,
			CuratedLocations:  curatedLocations,
		}
		if p.Status.StatusCode != domain.Success && p.Status.StatusCode != "" {
			contributorsResponse.ErrorMessage = &p.Status.Message
			contributorsResponse.ErrorCode = domain.StatusCodeToErrorCode(p.Status.StatusCode)
		}
		response.Purls = append(response.Purls, contributorsResponse)
	}
	return &response, nil
}

// convertOriginOutput converts an internal Provenance Output structure into a Provenance Response struct.
func convertOriginOutput(s *zap.SugaredLogger, output dtos.OriginOutput) (*pb.OriginResponse, error) {
	response := pb.OriginResponse{}
	for _, p := range output.Provenance {
		data, err := json.Marshal(p.Countries)
		if err != nil {
			s.Errorf("Problem marshalling origin country info for %s: %v", p.Purl, err)
			return &pb.OriginResponse{}, errors.New("problem marshalling origin country info")
		}
		var locations []*pb.Location
		err = json.Unmarshal(data, &locations)
		if err != nil {
			s.Errorf("Problem unmarshalling origin country info for %s: %v", p.Purl, err)
			return &pb.OriginResponse{}, errors.New("problem unmarshalling origin country info")
		}
		originResponse := &pb.OriginResponse_Purls{
			Purl:      p.Purl,
			Locations: locations,
		}
		if p.Status.StatusCode != domain.Success && p.Status.StatusCode != "" {
			originResponse.ErrorMessage = &p.Status.Message
			originResponse.ErrorCode = domain.StatusCodeToErrorCode(p.Status.StatusCode)
		}
		response.Purls = append(response.Purls, originResponse)
	}
	return &response, nil
}

// componentsRequestToDTO converts a components request into an internal ComponentDTO.
func componentsRequestToDTO(request *common.ComponentsRequest) ([]componenthelper.ComponentDTO, error) {
	if len(request.Components) == 0 {
		return []componenthelper.ComponentDTO{}, se.NewBadRequestError("Empty components provided", nil)
	}
	var componentDTOS []componenthelper.ComponentDTO
	emptyPurlsCount := 0
	for _, c := range request.Components {
		if c.Purl == "" {
			emptyPurlsCount++
			continue
		}
		componentDTOS = append(componentDTOS, componenthelper.ComponentDTO{
			Purl:        c.Purl,
			Requirement: c.Requirement,
		})
	}
	if emptyPurlsCount > 0 && emptyPurlsCount == len(request.Components) {
		return []componenthelper.ComponentDTO{}, se.NewBadRequestError("Empty purls provided", nil)
	}
	return componentDTOS, nil
}

// componentRequestToDTO converts a component request into an internal ComponentDTO.
func componentRequestToDTO(request *common.ComponentRequest) ([]componenthelper.ComponentDTO, error) {
	if request == nil || request.Purl == "" {
		return []componenthelper.ComponentDTO{}, se.NewBadRequestError("Empty component provided", nil)
	}
	return []componenthelper.ComponentDTO{
		{
			Purl:        request.Purl,
			Requirement: request.Requirement,
		},
	}, nil
}

// toComponentsContributorResponse converts an internal Provenance Output structure into a Provenance Response struct.
func toComponentsContributorResponse(output dtos.ProvenanceOutput) (*pb.ComponentsContributorResponse, error) {
	response := pb.ComponentsContributorResponse{
		ComponentsLocations: make([]*pb.ComponentLocationInfo, len(output.Provenance)),
	}
	for i, p := range output.Provenance {
		curatedData, err := json.Marshal(p.CuratedLocations)
		if err != nil {
			return &pb.ComponentsContributorResponse{}, errors.New("problem marshalling curated locations")
		}
		var curatedLocations []*pb.CuratedLocation
		err = json.Unmarshal(curatedData, &curatedLocations)
		if err != nil {
			return &pb.ComponentsContributorResponse{}, errors.New("problem unmarshalling curated locations")
		}

		declaredData, err := json.Marshal(p.DeclaredLocations)
		if err != nil {
			return &pb.ComponentsContributorResponse{}, errors.New("problem marshalling declared locations")
		}
		var declaredLocations []*pb.DeclaredLocation
		err = json.Unmarshal(declaredData, &declaredLocations)
		if err != nil {
			return &pb.ComponentsContributorResponse{}, errors.New("problem unmarshalling declared locations")
		}

		componentLocation := &pb.ComponentLocationInfo{
			Purl:              p.Purl,
			CuratedLocations:  curatedLocations,
			DeclaredLocations: declaredLocations,
		}
		if p.Status.StatusCode != domain.Success && p.Status.StatusCode != "" {
			componentLocation.ErrorMessage = &p.Status.Message
			componentLocation.ErrorCode = domain.StatusCodeToErrorCode(p.Status.StatusCode)
		}

		response.ComponentsLocations[i] = componentLocation
	}
	return &response, nil
}

// toComponentContributorResponse converts an internal Provenance Output structure into a Provenance Response struct.
func toComponentContributorResponse(output dtos.ProvenanceOutput) (*pb.ComponentContributorResponse, error) {
	response := pb.ComponentContributorResponse{
		ComponentLocations: &pb.ComponentLocationInfo{},
	}
	componentsContributors, err := toComponentsContributorResponse(output)
	if err != nil {
		return &response, se.NewInternalError("Error converting provenance data to response", err)
	}
	if len(componentsContributors.ComponentsLocations) > 0 {
		response.ComponentLocations = componentsContributors.ComponentsLocations[0]
	}
	return &response, nil
}

// toComponentsOriginResponse converts an internal Provenance Output structure into a Provenance Response struct.
func toComponentsOriginResponse(output dtos.OriginOutput) (*pb.ComponentsOriginResponse, error) {
	response := pb.ComponentsOriginResponse{
		ComponentsLocations: make([]*pb.ComponentLocation, len(output.Provenance)),
	}
	for i, p := range output.Provenance {
		data, err := json.Marshal(p.Countries)
		if err != nil {
			return &pb.ComponentsOriginResponse{}, errors.New("problem marshalling origin country info")
		}
		var locations []*pb.Location
		err = json.Unmarshal(data, &locations)
		if err != nil {
			return &pb.ComponentsOriginResponse{}, errors.New("problem unmarshalling origin country info")
		}
		componentLocation := &pb.ComponentLocation{
			Purl:      p.Purl,
			Locations: locations,
		}
		if p.Status.StatusCode != domain.Success && p.Status.StatusCode != "" {
			componentLocation.ErrorMessage = &p.Status.Message
			componentLocation.ErrorCode = domain.StatusCodeToErrorCode(p.Status.StatusCode)
		}
		response.ComponentsLocations[i] = componentLocation
	}
	return &response, nil
}

// toComponentsOriginResponse converts an internal Provenance Output structure into a Provenance Response struct.
func toComponentOriginResponse(output dtos.OriginOutput) (*pb.ComponentOriginResponse, error) {
	response := pb.ComponentOriginResponse{
		ComponentLocations: &pb.ComponentLocation{},
	}
	componentsOriginsResponse, err := toComponentsOriginResponse(output)
	if err != nil {
		return &response, se.NewInternalError("Error provenance data to response", err)
	}
	if len(componentsOriginsResponse.ComponentsLocations) > 0 {
		response.ComponentLocations = componentsOriginsResponse.ComponentsLocations[0]
	}
	return &response, nil
}
