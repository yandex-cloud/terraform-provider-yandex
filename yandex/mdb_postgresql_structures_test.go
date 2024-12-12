package yandex

import "testing"

func TestComparePGNoNamedHostInfo(t *testing.T) {
	tests := []struct {
		name        string
		existedHost *pgHostInfo
		newHost     *pgHostInfo
		expected    bool
	}{
		{
			name: "equal zone and subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11",
			},
			expected: true,
		},
		{
			name: "not equal zone and equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z12", subnetID: "sn11",
			},
			expected: false,
		},
		{
			name: "equal zone and not equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn12",
			},
			expected: false,
		},
		{
			name: "equal zone and empty new subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := matchesPGNoNamedHostInfo(tt.existedHost, tt.newHost); result != tt.expected {
				t.Errorf("matchesPGNoNamedHostInfo() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComparePGNamedHostInfo(t *testing.T) {
	tests := []struct {
		name        string
		existedHost *pgHostInfo
		newHost     *pgHostInfo
		expected    int
	}{
		{
			name: "equal zone and subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", name: "n1",
			},
			expected: 2,
		},
		{
			name: "not equal zone and equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z12", subnetID: "sn11", name: "n1",
			},
			expected: 0,
		},
		{
			name: "equal zone and not equal subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "sn12", name: "n1",
			},
			expected: 0,
		},
		{
			name: "equal zone and empty new subnetID",
			existedHost: &pgHostInfo{
				zone: "z11", subnetID: "sn11", fqdn: "fq11",
			},
			newHost: &pgHostInfo{
				zone: "z11", subnetID: "", name: "n1",
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := comparePGNamedHostInfo(tt.existedHost, tt.newHost, map[string]string{}); result != tt.expected {
				t.Errorf("comparePGNamedHostInfo() = %v, want %v", result, tt.expected)
			}
		})
	}
}
