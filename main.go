package main

import (
	"fmt"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/pedersandvoll/lighthouse/internal/config"
	systemd "github.com/pedersandvoll/lighthouse/internal/sytemd"
)

func main() {
	var conf config.Config
	refreshChan := make(chan struct{})
	err := config.Get(&conf, refreshChan)
	if err != nil {
		panic(err)
	}

	s, err := systemd.Connect()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	units := make(map[string]*dbus.UnitStatus)
	errChan := make(chan error)
	duration := time.Second * 2
	s.SubscribeUnits(&units, errChan, refreshChan, duration, &conf.Services)
	s.UnitsCleanup(&units, &conf.Services, duration/2)

	go func() {
		for err := range errChan {
			fmt.Printf("Error: %v\n", err)
		}
	}()
	time.Sleep(2 * time.Second)
	for {
		fmt.Println("--- Units ---")
		for _, unit := range units {
			fmt.Printf("Name: %s | Status: %s\n", unit.Name, unit.ActiveState)
		}
		time.Sleep(duration)
	}
}
