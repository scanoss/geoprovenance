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

package models

import (
	"testing"
)

func TestContributorOrigin(t *testing.T) {
	for name, load := range TestDataSets {
		t.Run(name, func(t *testing.T) {
			ctx, s, model := setupTestModel(t, load)
			list, err := model.GetTimeZoneOriginByPurlName(ctx, s, "scanoss/engine")
			if err != nil {
				t.Fatalf("unexpected error on model request: %v", err)
			}
			// Contributors without a timezone based country ('' or NULL) must not be counted
			got := map[string]int{}
			for _, o := range list {
				got[o.CountryName] = o.ContributorCount
			}
			want := map[string]int{"?": 1, "AR": 1, "BR": 1, "CO": 1}
			if len(got) != len(want) {
				t.Fatalf("expected %v, got %v", want, got)
			}
			for k, v := range want {
				if got[k] != v {
					t.Errorf("expected %v, got %v", want, got)
				}
			}
		})
	}
}
