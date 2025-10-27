package main

import (
	"fmt"

	"github.com/pedersandvoll/lighthouse/internal/config"
	systemd "github.com/pedersandvoll/lighthouse/internal/sytemd"
)

func main() {
	config, err := config.Get()
	if err != nil {
		panic(err)
	}

	s, err := systemd.Connect()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	for _, unit := range config.Services {
		us, err := s.GetUnitStatus(unit.SystemdUnit)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Name: %s | Status: %s\n", us.Name, us.ActiveState)
	}
}
