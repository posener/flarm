package process

import (
	"log"
	"math"

	"github.com/posener/flarm/flarmport"
)

type Processor struct {
	Lat, Long float64
	Alt       float64
}

type Object struct {
	ID        string
	Lat, Long float64
	// Altitude in m
	Alt float64
	// Ground speed in m/s
	GroundSpeed int64
	// Climb rate in m/s
	Climb float64
	Type  string
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
	lat, long := add(s.Lat, s.Long, e.RelativeNorth, e.RelativeEast)
	return &Object{
		Lat:         lat,
		Long:        long,
		Alt:         s.Alt + float64(e.RelativeVertical),
		Climb:       e.ClimbRate,
		GroundSpeed: e.GroundSpeed,
		Type:        e.AircraftType,
		ID:          id,
	}
}

func add(lat, lng float64, relN, relE int64) (float64, float64) {
	const earthRadius = 6378137

	dr := math.Sqrt(float64(relN*relN+relE*relE)) / earthRadius

	lat1 := (lat * (math.Pi / 180.0))
	lng1 := (lng * (math.Pi / 180.0))

	lat2_part1 := math.Sin(lat1) * math.Cos(dr)
	lat2_part2 := math.Cos(lat1) * math.Sin(dr) * float64(relE)

	lat2 := math.Asin(lat2_part1 + lat2_part2)

	lng2_part1 := float64(relN) * math.Sin(dr) * math.Cos(lat1)
	lng2_part2 := math.Cos(dr) - (math.Sin(lat1) * math.Sin(lat2))

	lng2 := lng1 + math.Atan2(lng2_part1, lng2_part2)
	lng2 = math.Mod((lng2+3*math.Pi), (2*math.Pi)) - math.Pi

	lat2 = lat2 * (180.0 / math.Pi)
	lng2 = lng2 * (180.0 / math.Pi)

	return lat2, lng2
}
