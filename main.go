package main

import (
	"fmt"
	"github.com/mkock/busproject/busservice"
)

func main() {
	fmt.Println("Starting simulation")
	expressLine := busservice.NewBus("Express Line")

	s1 := busservice.BusStop{Name: "Downtown"}
	s2 := busservice.BusStop{Name: "The University"}
	s3 := busservice.BusStop{Name: "The Village"}

	expressLine.AddStop(&s1)
	expressLine.AddStop(&s2)
	expressLine.AddStop(&s3)

	john := busservice.Prospect{
		SSN:         "12345612-22",
		Destination: &s2,
	}
	betty := busservice.Prospect{
		SSN:         "11223322-67",
		Destination: &s3,
	}
	s1.NotifyProspectArrival(john)
	s1.NotifyProspectArrival(betty)

	for expressLine.Go() {
		expressLine.VisitPassengers(func(p busservice.Passenger) {
			fmt.Printf("    Passenger with SSN %q is heading to %q\n", p.SSN, p.Destination.Name)
		})
	}
	fmt.Println("Simulation done")
}
