package yandex

import "testing"

func TestComparePGNoNamedHostInfo(t *testing.T) {
	if !matchesPGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "sn11",
	}) {
		t.Error("Hosts with equal zone and subnetID should match")
	}

	if matchesPGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z12", subnetID: "sn11",
	}) {
		t.Error("Host with not equal zone and equal subnetID should not match")
	}

	if matchesPGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "sn12",
	}) {
		t.Error("Host with equal zone and not equal subnetID should not match")
	}

	if !matchesPGNoNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "",
	}) {
		t.Error("Host with equal zone and empty new subnetID should match")
	}
}

func TestComparePGNamedHostInfo(t *testing.T) {
	if comparePGNamedHostInfo(&pgHostInfo{
		zone: "z11", subnetID: "sn11", fqdn: "fq11",
	}, &pgHostInfo{
		zone: "z11", subnetID: "sn11", name: "n1",
	}, map[string]string{}, "mhN") != 5 {
		t.Error("Compare host with equal zone and subnetID should return 5")
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
	}, map[string]string{}, "mhN") != 5 {
		t.Error("Compare host with equal zone and empty new subnetID should return 5")
	}
}
