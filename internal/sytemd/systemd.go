package systemd

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

type SystemdClient struct {
	Conn *dbus.Conn
	Ctx  context.Context
}

func Connect() (*SystemdClient, error) {
	ctx := context.Background()
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not connect to systemd: %w", err)
	}
	return &SystemdClient{Conn: conn, Ctx: ctx}, nil
}

func (s *SystemdClient) Close() {
	s.Conn.Close()
}

func (s *SystemdClient) GetUnitStatus(unit string) (*dbus.UnitStatus, error) {
	units, err := s.Conn.ListUnitsByNamesContext(s.Ctx, []string{unit})
	if err != nil {
		return nil, fmt.Errorf("failed to get unit status: %w", err)
	}
	if len(units) == 0 {
		return nil, fmt.Errorf("unit not found: %s", unit)
	}
	return &units[0], nil
}
