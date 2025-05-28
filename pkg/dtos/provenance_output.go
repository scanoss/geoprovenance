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

package dtos

type ProvenanceOutput struct {
	Provenance []ProvenanceOutputItem `json:"purls"`
}

type ProvenanceOutputItem struct {
	Purl              string                   `json:"purl"`
	DeclaredLocations []DeclaredProvenanceItem `json:"declared_locations"`
	CuratedLocations  []CuratedProvenanceItem  `json:"curated_locations"`
}

type DeclaredProvenanceItem struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type CuratedProvenanceItem struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}
