package process

import (
	"log"
	"math"
	"time"

	"github.com/posener/flarm/flarmport"
)

type Processor struct {
	Lat, Long float64
	Alt       float64
	TimeZone  *time.Location
	IDMap     map[string]string
}

type Object struct {
	ID        string
	Lat, Long float64
	// Direction of airplane (In degrees relative to N)
	Dir int
	// Altitude in m
	Alt float64
	// Ground speed in m/s
	GroundSpeed int64
	// Climb rate in m/s
	Climb      float64
	Type       string
	Time       time.Time
	AlarmLevel int
}

func (s Processor) Process(v interface{}) *Object {
	if v == nil {
		return nil
	}
	switch e := v.(type) {
	case flarmport.TypePFLAA:
		return s.processPFLAA(e)
	}
	return nil
}

func (s Processor) processPFLAA(e flarmport.TypePFLAA) *Object {
	id := e.ID
	if id == "" {
		log.Println("Ignoring empty ID entry.")
		return nil
	}
	// Apply ID mapping.
	if mappedID := s.IDMap[id]; mappedID != "" {
		id = mappedID
	}
	lat, long := add(s.Lat, s.Long, float64(e.RelativeNorth), float64(e.RelativeEast))
	return &Object{
		ID:          id,
		Lat:         lat,
		Long:        long,
		Dir:         int(e.Track),
		Alt:         s.Alt + float64(e.RelativeVertical),
		Climb:       e.ClimbRate,
		GroundSpeed: e.GroundSpeed,
		Type:        e.AircraftType,
		AlarmLevel:  int(e.AlarmLevel),
		Time:        time.Now().In(s.TimeZone),
	}
}

func add(lat, lon float64, relN, relE float64) (float64, float64) {
	const earthRadius = 6378137

	//Coordinate offsets in radians
	dLat := relN / earthRadius
	dLon := relE / (earthRadius * math.Cos(math.Pi*lat/180.0))

	//OffsetPosition, decimal degrees
	return lat + dLat*180.0/math.Pi,
		lon + dLon*180.0/math.Pi
}
