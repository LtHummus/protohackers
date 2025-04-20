package speeding

import (
	"bytes"
	"encoding/binary"
	"math"
	"sort"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	MessageKindError         = 0x10
	MessageKindPlate         = 0x20
	MessageKindTicket        = 0x21
	MessageKindWantHeartbeat = 0x40
	MessageKindHeartbeat     = 0x41
	MessageKindIAmCamera     = 0x80
	MessageKindIAmDispatch   = 0x81
)

type System struct {
	limitLock       *sync.Mutex
	observationLock *sync.Mutex
	daysLock        *sync.Mutex
	roadChannelLock *sync.Mutex

	speedLimits  map[uint16]float64
	observations map[string]map[uint16][]*Observation

	ticketDays map[string]map[uint32]bool

	roadChannels map[uint16]chan []byte
}

func NewSystem() *System {
	return &System{
		limitLock:       &sync.Mutex{},
		observationLock: &sync.Mutex{},
		daysLock:        &sync.Mutex{},
		roadChannelLock: &sync.Mutex{},
		speedLimits:     map[uint16]float64{},
		observations:    map[string]map[uint16][]*Observation{},
		ticketDays:      map[string]map[uint32]bool{},
		roadChannels:    map[uint16]chan []byte{},
	}
}

func (s *System) RegisterObservation(plate string, road uint16, timestamp uint32, mileMarker uint16) {
	s.observationLock.Lock()
	defer s.observationLock.Unlock()

	if s.observations[plate] == nil {
		s.observations[plate] = map[uint16][]*Observation{}
	}

	if s.observations[plate][road] == nil {
		s.observations[plate][road] = make([]*Observation, 0)
	}

	log.Info().Str("plate", plate).Uint16("road", road).Uint32("timestamp", timestamp).Msg("registered observation")
	s.observations[plate][road] = append(s.observations[plate][road], &Observation{
		timestamp:  timestamp,
		mileMarker: mileMarker,
	})
}

func (s *System) checkForTickets(plate string, road uint16) {
	log.Info().Str("plate", plate).Uint16("road", road).Msg("checking for tickets")
	limit := s.speedLimits[road]
	if limit == 0 {
		log.Warn().Uint16("road", road).Msg("tried to check for tickets with no speed limit")
		return
	}

	observations := s.observations[plate][road]

	sort.Slice(observations, func(i, j int) bool {
		return observations[i].IsBefore(observations[j])
	})

	// for now, we only check consecutive observations, which should be enough?
	for i := 0; i < len(observations)-1; i++ {
		a := observations[i]
		b := observations[i+1]

		speed := a.SpeedBetween(b)

		diff := speed - limit

		// this may need to be tweaked
		if diff > 0.1 {
			officialSpeed := uint16(math.Round(speed * 100))
			log.Info().Uint16("road", road).Uint16("speed", officialSpeed).Str("plate", plate).Msg("issuing ticket")
			s.issueTicket(plate, road, a.mileMarker, a.timestamp, b.mileMarker, b.timestamp, officialSpeed)
		}
	}

}

func (s *System) issueTicket(plate string, road uint16, mile1 uint16, timestamp1 uint32, mile2 uint16, timestamp2 uint32, speed uint16) {
	if s.hasIssuedTicket(plate, timestamp1) || s.hasIssuedTicket(plate, timestamp2) {
		log.Warn().Str("plate", plate).Uint32("day_num_1", dayNum(timestamp1)).Uint32("day_num_2", dayNum(timestamp2)).Msg("already issued ticket for day")
		return
	}

	s.recordTicketDay(plate, timestamp1)
	s.recordTicketDay(plate, timestamp2)

	buf := &bytes.Buffer{}

	buf.WriteByte(MessageKindTicket)
	WriteString(buf, plate)

	remainingTicketDetails := struct {
		Road       uint16
		Mile1      uint16
		Timestamp1 uint32
		Mile2      uint16
		Timestamp2 uint32
		Speed      uint16
	}{
		road, mile1, timestamp1, mile2, timestamp2, speed,
	}

	binary.Write(buf, binary.BigEndian, &remainingTicketDetails)

	payload := buf.Bytes()

	log.Debug().Hex("payload", payload).Msg("generated ticket")

	// guaranteed to exist at this point since it's created with the camera
	s.roadChannels[road] <- payload
}

func dayNum(timestamp uint32) uint32 {
	return timestamp / 86400
}

func (s *System) hasIssuedTicket(plate string, timestamp uint32) bool {
	s.daysLock.Lock()
	defer s.daysLock.Unlock()

	plateMap := s.ticketDays[plate]
	if plateMap == nil {
		s.ticketDays[plate] = map[uint32]bool{}
	}

	return s.ticketDays[plate][dayNum(timestamp)]
}

func (s *System) recordTicketDay(plate string, timestamp uint32) {
	s.daysLock.Lock()
	defer s.daysLock.Unlock()

	s.ticketDays[plate][dayNum(timestamp)] = true
}

func (s *System) RecordObservation(plate string, road uint16, mile uint16, timestamp uint32) {
	s.observationLock.Lock()
	defer s.observationLock.Unlock()

	obsMap := s.observations[plate]
	if obsMap == nil {
		s.observations[plate] = map[uint16][]*Observation{}
	}

	roadMap := obsMap[road]
	if roadMap == nil {
		s.observations[plate][road] = []*Observation{}
	}

	s.observations[plate][road] = append(s.observations[plate][road], &Observation{
		timestamp:  timestamp,
		mileMarker: mile,
	})

	log.Info().Str("plate", plate).Uint16("road", road).Uint16("mile", mile).Uint32("timestamp", timestamp).Msg("recorded observation")

	s.checkForTickets(plate, road)

}

func (s *System) RegisterRoad(road uint16, limit uint16) {
	s.roadChannelLock.Lock()
	defer s.roadChannelLock.Unlock()

	if s.speedLimits[road] != 0 {
		log.Warn().Uint16("road", road).Msg("already have a limit for this road")
	} else {
		s.speedLimits[road] = float64(limit)
	}

	if s.roadChannels[road] != nil {
		log.Warn().Uint16("road", road).Msg("already have a channel for this road")
	} else {
		s.roadChannels[road] = make(chan []byte, 1000)
	}

	log.Info().Uint16("road", road).Uint16("limit", limit).Msg("registered road")
}
