package cmd

import (
	"fmt"
	"strings"

	"github.com/Zhima-Mochi/autovpn/pritunl"
)

// VPNManager interface for abstraction of VPN tool management
type VPNManager interface {
	Profiles() ([]pritunl.Profile, error)
	Connections() (map[string]pritunl.Connection, error)
	List(profiles []pritunl.Profile, connections map[string]pritunl.Connection) error
	Connect(profileID string) error
	Disconnect(profileID string) error
}

// GetVPNManager returns the appropriate VPN manager based on the tool specified
func GetVPNManager(tool string) (VPNManager, error) {
	switch strings.ToLower(tool) {
	case "pritunl":
		return pritunl.GetPritunlManager(), nil
	default:
		return nil, fmt.Errorf("unsupported VPN tool: %s", tool)
	}
}
