package busservice

import (
	"fmt"
	"strconv"
	"time"
)

// SeniorAge is the minimum age from which a Passenger is considered a senior to the BusCompany.
const SeniorAge = 65

// Passenger represents a bus passenger, uniquely identified by their SSN.
type Passenger struct {
	SSN            string
	SeatNumber     uint8
	Destination    *BusStop
	HasValidTicket bool
}

// Charge prints a message that the Passenger has been charged "amount" money, and returns a copy with validTicket = true.
func (p Passenger) Charge(amount float64) Passenger {
	if p.HasValidTicket {
		return p // We already charged this Passenger.
	}
	fmt.Printf("Passenger with SSN %s: charged %.2f of arbitrary money\n", p.SSN, amount)
	p.HasValidTicket = true
	return p
}

// IsSenior returns true if the Passenger is a senior, and false otherwise.
// IsSenior detects age by extracting the last two digits from the SSN and treating them like an age.
func (p Passenger) IsSenior() bool {
	age, err := strconv.ParseInt(p.SSN[len(p.SSN)-2:], 10, 8)
	if err != nil {
		panic("invalid SSN: " + p.SSN)
	}
	return age >= SeniorAge
}

// Passengers represents a set of Passengers, using their SSN as key.
type Passengers map[string]Passenger

// NewPassengerSet returns an empty set of Passengers, ready to use.
func NewPassengerSet() Passengers {
	return make(map[string]Passenger)
}

// Visit calls visitor once for every Passenger in the set.
func (p Passengers) Visit(visitor func(Passenger)) {
	for _, one := range p {
		visitor(one)
	}
}

// Find returns the Passenger with the given SSN. If none was found, an empty Passenger is returned.
func (p Passengers) Find(ssn string) Passenger {
	if one, ok := p[ssn]; ok {
		return one
	}
	return Passenger{}
}

// VisitUpdate calls visitor for each Passenger in the set. Updating their SSN's is not recommended.
func (p *Passengers) VisitUpdate(visitor func(*Passenger)) {
	for ssn, pp := range *p {
		visitor(&pp)
		(*p)[ssn] = pp
	}
}

// Manifest returns the SSN's of all Passengers in the set.
func (p Passengers) Manifest() []string {
	ssns := make([]string, 0, len(p))
	p.Visit(func(p Passenger) { ssns = append(ssns, p.SSN) })
	return ssns
}

// Bus carries Passengers from A to B if they have a valid bus ticket.
type Bus struct {
	Company     BusCompany
	name        string
	passengers  Passengers
	stops       []*BusStop
	currentStop int16
}

// NewBus returns a new Bus with an empty passenger set.
func NewBus(name string) Bus {
	b := Bus{}
	b.name = name
	b.currentStop = -1
	b.passengers = NewPassengerSet()
	return b
}

// add adds a single Passenger to the Bus. For brevity, we don't care too much about accidentally adding the same Passenger more than once.
func (b *Bus) add(p Passenger) {
	if b.passengers == nil {
		b.passengers = make(map[string]Passenger)
	}
	b.passengers[p.SSN] = p
	fmt.Printf("%s: boarded passenger with SSN %q\n", b.name, p.SSN)
}

// Board adds the given Passenger to the Bus and charges them a ticket price calculated by chargeFn if they don't already have a paid ticket.
// Board returns false if the Passenger was not allowed to board the Bus.
func (b *Bus) Board(p *Passenger, chargeFn PriceCalculator) bool {
	var allowed bool // Default value is false
	if p.HasValidTicket {
		allowed = true
	} else {
		amount := chargeFn(*p)
		p2 := p.Charge(amount)
		p = &p2
		allowed = true
	}
	if allowed {
		b.add(*p)
	}
	return allowed
}

// Remove removes a single Passenger from the Bus.
func (b *Bus) Remove(p Passenger) {
	delete(b.passengers, p.SSN)
	fmt.Printf("%s: unboarded passenger with SSN %q\n", b.name, p.SSN)
}

// AddStop adds the given BusStop to the list of stops that the Bus will stop at. Each stop is visited in order.
func (b *Bus) AddStop(busStop *BusStop) {
	b.stops = append(b.stops, busStop)
}

// Go takes the Bus to the next BusStop. Go returns true if there are still more stops to visit.
func (b *Bus) Go() bool {
	b.currentStop++
	lastIndex := int16(len(b.stops) - 1)
	if b.currentStop == lastIndex {
		fmt.Printf("%s: reached the end of the line, everybody out\n", b.name)
		b.VisitPassengers(func(p Passenger) {
			b.Remove(p)
		})
		return false
	}
	if b.currentStop == 0 {
		fmt.Printf("%s: starting\n", b.name)
	} else {
		fmt.Printf("%s: carrying %d passengers: heading for next stop\n", b.name, len(b.passengers))
	}
	curr := b.stops[b.currentStop]
	fmt.Printf("%s: arriving at %q\n", b.name, curr.Name)
	curr.NotifyBusArrival(b)
	return b.currentStop < lastIndex
}

// Manifest asks Passengers for a SSN manifest and returns it.
func (b Bus) Manifest() []string {
	return b.passengers.Manifest()
}

// VisitPassengers calls function visitor for each Passenger on the bus.
func (b *Bus) VisitPassengers(visitor func(Passenger)) {
	b.passengers.Visit(visitor)
}

// FindPassenger returns the Passenger that matches the given SSN, if found. Otherwise, an empty Passenger is returned.
func (b *Bus) FindPassenger(ssn string) Passenger {
	if p, ok := b.passengers[ssn]; ok {
		return p
	}
	return Passenger{} // A nobody.
}

// UpdatePassengers calls function visitor for each Passenger on the bus. Passengers are passed by reference and may be modified.
func (b *Bus) UpdatePassengers(visitor func(*Passenger)) {
	ps := make(map[string]Passenger, len(b.passengers))
	for ssn, p := range b.passengers {
		visitor(&p)
		ps[ssn] = p
	}
	b.passengers = ps
}

// NotifyBoardingIntent is called by BusStop every time a Prospect arrives and instructs the Bus to signal its arrival when at that BusStop.
func (b *Bus) NotifyBoardingIntent(busStop *BusStop) {
	if b.StopsAt(busStop) {
		return // We already intend to stop here.
	}
	b.AddStop(busStop)
}

// NotifyArrival notifies the current BusStop that the Bus has arrived.
func (b *Bus) NotifyArrival() {
	curr := b.stops[b.currentStop]
	curr.NotifyBusArrival(b)
}

// StopsAt checks if Bus stops at the given BusStop, and returns true if it does, and false otherwise.
func (b Bus) StopsAt(busStop *BusStop) bool {
	for _, stop := range b.stops {
		if stop.Equals(busStop) {
			return true
		}
	}
	return false
}

// CurrentStop returns the BusStop that the Bus is currently stopped at.
func (b Bus) CurrentStop() *BusStop {
	return b.stops[b.currentStop]
}

// Prospect is a potential Passenger. Prospects wait at BusStops to board Buses.
type Prospect struct {
	SSN         string
	Destination *BusStop
}

// ToPassenger returns a Passenger with the same SSN as his or her Prospect.
func (p Prospect) ToPassenger() Passenger {
	return Passenger{SSN: p.SSN, Destination: p.Destination}
}

// BusStop represents a place where a Bus can stop and signal to prospects (future passengers)
// that they may board.
type BusStop struct {
	Name      string
	prospects []Prospect
	busses    []Bus
}

// Equals returns true if the given BusStop is the same as the receiver.
func (b *BusStop) Equals(busStop *BusStop) bool {
	return b.Name == busStop.Name
}

// NotifyBusArrival is called by Bus upon arrival.
func (b *BusStop) NotifyBusArrival(bus *Bus) {
	bus.VisitPassengers(func(p Passenger) {
		if bus.CurrentStop().Equals(p.Destination) {
			bus.Remove(p)
		}
	})
	for _, p := range b.prospects {
		if bus.StopsAt(p.Destination) {
			pas := p.ToPassenger()
			bus.Board(&pas, bus.Company.GetPricing())
		}
	}
}

// NotifyProspectArrival is called whenever a prospect arrives at Busstop.
func (b *BusStop) NotifyProspectArrival(p Prospect) {
	b.prospects = append(b.prospects, p)

	// Find all Busses on this route.
	for _, bus := range b.busses {
		if bus.StopsAt(p.Destination) {
			bus.NotifyBoardingIntent(b)
		}
	}
}

// WorkdayPricing charges EUR 6 for regular Passengers and EUR 4.5 for seniors during workdays.
func WorkdayPricing(p Passenger) float64 {
	if p.IsSenior() {
		return 4.5
	}
	return 6.0
}

// WeekendPricing charges EUR 5 for regular Passengers and EUR 3.5 for seniors during weekends.
func WeekendPricing(p Passenger) float64 {
	if p.IsSenior() {
		return 3.5
	}
	return 5.0
}

// PriceCalculator is the type used by BusCompany to determine the ticket price for a Passenger.
// PriceCalculator returns the ticket price in the local currency.
type PriceCalculator func(p Passenger) float64

// BusCompany represents the bus company responsible for the Bus service. BusCompany determines price policies.
type BusCompany string

// GetPricing returns a price calculator based on the pricing policy of the day.
func (b BusCompany) GetPricing() PriceCalculator {
	wd := time.Now().Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return WeekendPricing
	}
	return WorkdayPricing
}
