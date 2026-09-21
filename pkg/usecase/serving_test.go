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

package usecase

import "testing"

func TestServingPurlKey(t *testing.T) {
	tests := []struct {
		purl    string
		want    string
		wantErr bool
	}{
		{purl: "pkg:npm/express", want: "pkg:npm/express"},
		{purl: "pkg:npm/express@4.18.2", want: "pkg:npm/express"},
		{purl: "pkg:github/scanoss/engine", want: "pkg:github/scanoss/engine"},
		{purl: "pkg:github/scanoss/engine@v5.0.0?foo=bar#sub/path", want: "pkg:github/scanoss/engine"},
		{purl: "pkg:golang/github.com/scanoss/engine", want: "pkg:github/scanoss/engine"},
		{purl: "pkg:npm/%40angular/core", want: "pkg:npm/@angular/core"},
		{purl: "", wantErr: true},
		{purl: "not-a-purl", wantErr: true},
	}
	for _, tt := range tests {
		got, err := ServingPurlKey(tt.purl)
		if (err != nil) != tt.wantErr {
			t.Errorf("ServingPurlKey(%q) error = %v, wantErr %v", tt.purl, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("ServingPurlKey(%q) = %q, want %q", tt.purl, got, tt.want)
		}
	}
}
