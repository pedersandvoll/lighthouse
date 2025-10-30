package systemd

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/pedersandvoll/lighthouse/internal/config"
)

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

func Init(timePtr *int, conf *config.Config, refreshChan *chan struct{}) (*SystemdClient, SystemdState, error) {
	s, err := Connect()
	if err != nil {
		return nil, SystemdState{}, fmt.Errorf("could not connect to systemd: %w", err)
	}

	units := make(map[string]*UnitStatus)
	errChan := make(chan error)
	duration := time.Second * time.Duration(*timePtr)

	s.SubscribeUnits(&units, errChan, *refreshChan, duration, &conf.Services)
	s.UnitsCleanup(&units, &conf.Services, duration/2)

	return s, SystemdState{ErrChan: &errChan, Units: &units, Duration: duration}, nil
}

func (s *SystemdClient) SubscribeUnits(resultMap *map[string]*UnitStatus, errChan chan<- error, refreshChan <-chan struct{}, duration time.Duration, units *[]config.Service) {
	unitStatuses, errs := s.Conn.SubscribeUnits(duration)

	go func() {
		for {
			select {
			case statuses := <-unitStatuses:
				for _, unit := range *units {
					if status, ok := statuses[unit.SystemdUnit]; ok {
						(*resultMap)[unit.Name] = &UnitStatus{UnitStatus: status, Available: true}
					} else {
						(*resultMap)[unit.Name] = &UnitStatus{UnitStatus: nil, Available: false}
					}
				}
			case <-refreshChan:
				for _, unit := range *units {
					status, _ := s.GetUnitStatus(unit.SystemdUnit)
					if status != nil {
						(*resultMap)[unit.Name] = &UnitStatus{UnitStatus: status, Available: true}
					} else {
						(*resultMap)[unit.Name] = &UnitStatus{UnitStatus: nil, Available: false}
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

func (s *SystemdClient) UnitsCleanup(unitsMap *map[string]*UnitStatus, units *[]config.Service, duration time.Duration) {
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		for range ticker.C {
			validUnits := make(map[string]struct{}, len(*units))
			for _, unit := range *units {
				validUnits[unit.Name] = struct{}{}
			}

			for name := range *unitsMap {
				if _, exists := validUnits[name]; !exists {
					delete(*unitsMap, name)
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
