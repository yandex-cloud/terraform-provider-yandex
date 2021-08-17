package yandex

import "testing"

func TestComparePGNoNamedHostInfo(t *testing.T) {
	if comparePGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "sn11",
	}, map[string]struct{}{}) != 1 {
		t.Error("Compare host with equal zone and subnetID should return 1")
	}

	if comparePGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z12", subnetID: "sn11",
	}, map[string]struct{}{}) != 0 {
		t.Error("Compare host with not equal zone and equal subnetID should return 0")
	}

	if comparePGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "sn12",
	}, map[string]struct{}{}) != 0 {
		t.Error("Compare host with equal zone and not equal subnetID should return 0")
	}

	if comparePGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "",
	}, map[string]struct{}{}) != 1 {
		t.Error("Compare host with equal zone and empty new subnetID should return 1")
	}
}

func TestComparePGNamedHostInfo(t *testing.T) {
	if comparePGNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "sn11", name: "n1",
	}, map[string]string{}, "mhN") != 4 {
		t.Error("Compare host with equal zone and subnetID should return 4")
	}

	if comparePGNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z12", subnetID: "sn11", name: "n1",
	}, map[string]string{}, "mhN") != 0 {
		t.Error("Compare host with not equal zone and equal subnetID should return 0")
	}

	if comparePGNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "sn12", name: "n1",
	}, map[string]string{}, "mhN") != 0 {
		t.Error("Compare host with equal zone and not equal subnetID should return 0")
	}

	if comparePGNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "", name: "n1",
	}, map[string]string{}, "mhN") != 4 {
		t.Error("Compare host with equal zone and empty new subnetID should return 4")
	}
}
