package systemd

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/pedersandvoll/lighthouse/internal/config"
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
	if err := conn.Subscribe(); err != nil {
		return nil, fmt.Errorf("could not subscribe to systemd dbus event")
	}

	return &SystemdClient{Conn: conn, Ctx: ctx}, nil
}

func (s *SystemdClient) Close() {
	s.Conn.Close()
}

func (s *SystemdClient) SubscribeUnits(resultMap *map[string]*dbus.UnitStatus, errChan chan<- error, refreshChan <-chan struct{}, duration time.Duration, units *[]config.Service) {
	unitStatuses, errs := s.Conn.SubscribeUnits(duration)

	go func() {
		for {
			select {
			case statuses := <-unitStatuses:
				for _, unit := range *units {
					if status, ok := statuses[unit.SystemdUnit]; ok {
						(*resultMap)[unit.Name] = status
					}
				}
			case <-refreshChan:
				for _, unit := range *units {
					status, _ := s.GetUnitStatus(unit.SystemdUnit)
					if status != nil {
						(*resultMap)[unit.Name] = status
					}
				}
			case err := <-errs:
				if errChan != nil {
					errChan <- err
				}
			}
		}
	}()
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
