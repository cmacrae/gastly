// Copyright 2019 Calum MacRae. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package gastly

import "testing"

func TestRandProxy(t *testing.T) {
	data := Provider{
		Data: []Container{
			// Container{
			// 	Proxy: Proxy{
			// 		IP:          "127.0.0.1",
			// 		PortNum:     "80",
			// 		CountryCode: "UK",
			// 		CountryName: "United Kingdom",
			// 		RegionName:  "England",
			// 		CityName:    "London",
			// 		Status:      "online",
			// 		PanelUser:   "",
			// 		PanelPass:   "",
			// 	},
			// },
			Container{
				Proxy: Proxy{
					IP:          "127.0.0.2",
					PortNum:     "81",
					CountryCode: "US",
					CountryName: "United States",
					RegionName:  "New York",
					CityName:    "New York",
					Status:      "offline",
					PanelUser:   "",
					PanelPass:   "",
				},
			},
		},
	}

	selection := data.RandProxy().CountryCode
	// FIXME: vet complains about this comparison...
	// if selection != "UK" || selection != "UK" {
	if selection != "US" {
		t.Errorf("RandProxy(%q).CountryCode == %q, want \"US\" or \"UK\"", data, selection)
	}
}
