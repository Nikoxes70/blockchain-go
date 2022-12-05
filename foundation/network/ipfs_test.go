package network

import (
	"testing"
)

func Test_GET_Host(t *testing.T) {
	adds := GetHost()
	if adds == "::1" {
		t.Errorf("WRONG ADDRESS ADDRESS - -> %v <-", adds)
	}
}
