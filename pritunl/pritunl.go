package pritunl

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cghdev/gotunl/pkg/gotunl"
	"github.com/fatih/color"
)

var (
	once     sync.Once
	instance *PritunlManager
)

type PritunlManager struct {
	gotunl  *gotunl.Gotunl
	servers map[string]*ServerInfo
}

func GetPritunlManager() *PritunlManager {
	once.Do(func() {
		instance = &PritunlManager{
			gotunl: gotunl.New(),
		}
	})
	return instance
}

type Profile struct {
	ID     string
	Path   string
	Server string
	User   string
}

type Conf struct {
	Name   string `json:"name"`
	Server string `json:"server"`
	User   string `json:"user"`
}

type Connection struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Timestamp  int64  `json:"timestamp"`
	ServerAddr string `json:"server_addr"`
	ClientAddr string `json:"client_addr"`
}

type ServerInfo struct {
	LastUsed          time.Time
	RunningTime       time.Duration
	OtherInfo         map[string]interface{}
	ConnectionStarted time.Time
}

// Profiles retrieves the list of available VPN profiles and sorts them by server order
func (pm *PritunlManager) Profiles() ([]Profile, error) {
	var profiles []Profile

	for id, profile := range pm.gotunl.Profiles {
		var conf Conf
		err := json.Unmarshal([]byte(profile.Conf), &conf)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile config for ID %s: %v", id, err)
		}

		profiles = append(profiles, Profile{
			ID:     id,
			Path:   profile.Path,
			Server: conf.Server,
			User:   conf.User,
		})
	}

	return profiles, nil
}

// Connections retrieves the current active connections
func (pm *PritunlManager) Connections() (map[string]Connection, error) {
	var conns map[string]Connection

	connStr := pm.gotunl.GetConnections()
	err := json.Unmarshal([]byte(connStr), &conns)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal connections: %v", err)
	}

	return conns, nil
}

// Connect attempts to connect to a VPN profile by its ID and password
func (pm *PritunlManager) Connect(id string) error {
	config, err := getConfig(id)
	if err != nil {
		color.Red("Failed to get config: %v", err)
		return err
	}

	isSucess := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				isSucess = false
			}
		}()
		pm.gotunl.ConnectProfile(id, "pritunl", config.OTP())
	}()

	if !isSucess {
		return errors.New("failed to connect to VPN")
	}
	return nil
}

// Disconnect disconnects from a VPN profile by its ID
func (pm *PritunlManager) Disconnect(id string) error {
	isSucess := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				isSucess = false
			}
		}()
		pm.gotunl.DisconnectProfile(id)
	}()

	if !isSucess {
		return errors.New("failed to disconnect from VPN")
	}
	return nil
}

// DisconnectAll disconnects all active VPN connections
func (pm *PritunlManager) DisconnectAll() {
	pm.gotunl.StopConnections()
}

func (pm *PritunlManager) List(profiles []Profile, conns map[string]Connection) error {
	if len(profiles) == 0 {
		return errors.New("no profile found in Pritunl")
	}

	var rows [][]string
	for i, profile := range profiles {
		row := []string{strconv.Itoa(i + 1), profile.Server, profile.User, "disconnected", "", "", ""}
		if conn, ok := conns[profile.ID]; ok {
			row[3] = conn.Status
			if conn.Timestamp > 0 {
				row[4] = formatDuration(time.Since(time.Unix(conn.Timestamp, 0)))
			}
			row[5] = conn.ClientAddr
			row[6] = conn.ServerAddr
		}
		row[3] = formatStatus(row[3])
		rows = append(rows, row)
	}

	table := createTable()
	table.AppendBulk(rows)
	table.Render()

	return nil
}

func formatStatus(status string) string {
	status = strings.ToUpper(status)
	switch status {
	case "CONNECTED":
		return color.New(color.FgGreen, color.Bold).SprintfFunc()(status)
	case "CONNECTING":
		return color.New(color.FgYellow, color.Bold).SprintfFunc()(status + "...")
	case "DISCONNECTING":
		return color.New(color.FgBlack, color.Bold).SprintfFunc()(status + "...")
	case "DISCONNECTED":
		return color.New(color.FgBlack, color.Bold).SprintfFunc()(status)
	default:
		return color.New(color.FgBlack, color.Bold).SprintfFunc()("UNKNOWN")
	}
}

func formatDuration(sec time.Duration) string {
	d := sec / (24 * time.Hour)
	h := (sec % (24 * time.Hour)) / time.Hour
	m := (sec % time.Hour) / time.Minute
	s := (sec % time.Minute) / time.Second

	parts := []string{}
	if d > 0 {
		parts = append(parts, fmt.Sprintf("%dd", d))
	}
	if h > 0 || d > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 || h > 0 || d > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}

	parts = append(parts, fmt.Sprintf("%ds", s))

	return strings.Join(parts, " ")
}
