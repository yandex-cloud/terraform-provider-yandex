package acceptance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSentinelHosts(t *testing.T) {
	diskTypeId := ""
	publicIPFlags := []*bool{nil}
	replicaPriorities := []*int{nil}
	expected := `
  host {
  	zone      = "ru-central1-d"
	subnet_id = "${yandex_vpc_subnet.foo.id}"
	
	
  }
`

	actual := getSentinelHosts(diskTypeId, publicIPFlags, replicaPriorities)
	require.Equal(t, expected, actual)
}
