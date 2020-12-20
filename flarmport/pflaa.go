package flarmport

import "github.com/adrianmo/go-nmea"

// PFLAA - Data on other proximate aircraft
// Syntax:
//   PFLAA,<AlarmLevel>,<RelativeNorth>,<RelativeEast>,
//   <RelativeVertical>,<IDType>,<ID>,<Track>,<TurnRate>,<GroundSpeed>,
//   <ClimbRate>,<AcftType>
//
// Description:
//
// Data on other proximate aircraft, intended for connected devices with sufficient
// CPU performance. This sentence should be treated with utmost flexibility and
// tolerance on a best effort base. Individual parameters may be empty. The
// sentence is only sent when port baud rate is 19.2k or higher. In case of serial
// port congestion or high CPU load, this sentence may be omitted for several
// objects independent of the alarm level. Non-directional targets (transponder
// Mode-C/S; protocol version 6 and higher) are output as PFLAA sentences.
//
// Obstacle information is not delivered with this sentence.
//
// Note that in case of many targets within range, individual targets, including the
// most dangerous one, might not be delivered every second, not regularly, or not
// at all, due to less strict priority handling for the PFLAA sentence. Always use
// PFLAU as primary alarm source. Usually, but not always, the last PFLAA
// sentence is the one causing the PFLAU content. The other PFLAA sentences are
// not ordered. Do not expect to receive PFLAU <Rx> times PFLAA sentences,
// because the number of aircraft being processed might be higher or lower. PFLAA
// sentences can be based on extrapolated historical data. PFLAA sentences are
// limited to other aircraft with a horizontal and vertical distance less than the
// configured range. On Classic FLARM, the vertical distance is always 500 m. Nonmoving aircraft are
// suppressed.
type TypePFLAA struct {
	nmea.BaseSentence `json:"-"`
	// Decimal integer value. Range: from 0 to 3.
	// Alarm level as assessed by FLARM:
	// 0 = no alarm (also used for no-alarm traffic information)
	// 1 = alarm, 13-18 seconds to impact
	// 2 = alarm, 9-12 seconds to impact
	// 3 = alarm, 0-8 seconds to impact
	AlarmLevel int64
	// Decimal integer value. Range: from -32768 to 32767.
	// Relative position in meters true north from own position. If
	// <RelativeEast> is empty, <RelativeNorth> represents the
	// estimated distance to a target with unknown bearing
	// (transponder Mode-C/S).
	RelativeNorth int64
	// Decimal integer value. Range: from -32768 to 32767.
	// Relative position in meters true east from own position.
	// Field is empty for non-directional targets.
	RelativeEast int64
	// Decimal integer value. Range: from -32768 to 32767.
	// Relative vertical separation in meters above own position.
	// Negative values indicate that the other aircraft is lower.
	// Some distance-dependent random noise is applied to
	// altitude data if stealth mode is activated either on the
	// target or own aircraft and no alarm is present at this time.
	RelativeVertical int64
	// Decimal integer value. Range: from 0 to 3.
	// Defines the interpretation of the following field <ID>
	// 1 = official ICAO 24-bit aircraft address
	// 2 = stable FLARM ID (chosen by FLARM)
	// 3 = anonymous ID, used if stealth mode is activated
	// either on the target or own aircraft
	// Field is empty if no identification is known (e.g. transponder
	// Mode-C).
	IDType string
	// 6-digit hexadecimal value (e.g. “5A77B1”) as configured in
	// the target’s PFLAC,,ID sentence. The interpretation is
	// delivered in <ID-Type>.
	// Field is empty if no identification is known (e.g.
	// Transponder Mode-C). Random ID will be sent if stealth
	// mode is activated either on the target or own aircraft and
	// no alarm is present at this time.
	ID string
	// Decimal integer value. Range: from 0 to 359.
	// The target’s true ground track in degrees. The value 0
	// indicates a true north track. This field is empty if stealth
	// mode is activated either on the target or own aircraft and
	// for non-directional targets.
	Track int64
	// Currently this field is empty.
	TurnRate int64
	// Decimal integer value. Range: from 0 to 32767.
	// The target’s ground speed in m/s. The field is 0 to indicate
	// that the aircraft is not moving, i.e. on ground. This field is
	// empty if stealth mode is activated either on the target or
	// own aircraft and for non-directional targets.
	GroundSpeed int64
	// Decimal fixed point number with one digit after the radix
	// point (dot). Range: from -32.7 to 32.7.
	// The target’s climb rate in m/s. Positive values indicate a
	// climbing aircraft. This field is empty if stealth mode is
	// activated either on the target or own aircraft and for nondirectional targets.
	ClimbRate float64
	// Hexadecimal value. Range: from 0 to F.
	// Aircraft types:
	// 0 = unknown
	// 1 = glider / motor glider
	// 2 = tow / tug plane
	// 3 = helicopter / rotorcraft
	// 4 = skydiver
	// 5 = drop plane for skydivers
	// 6 = hang glider (hard)
	// 7 = paraglider (soft)
	// 8 = aircraft with reciprocating engine(s)
	// 9 = aircraft with jet/turboprop engine(s)
	// A = unknown
	// B = balloon
	// C = airship
	// D = unmanned aerial vehicle (UAV)
	// E = unknown
	// F = static object
	AircraftType string
}

func init() {
	nmea.MustRegisterParser("FLAA", func(s nmea.BaseSentence) (nmea.Sentence, error) {
		// This example uses the package builtin parsing helpers
		// you can implement your own parsing logic also
		p := nmea.NewParser(s)
		return TypePFLAA{
			BaseSentence:     s,
			AlarmLevel:       p.Int64(0, "alarm level"),
			RelativeNorth:    p.Int64(1, "rel north"),
			RelativeEast:     p.Int64(2, "rel east"),
			RelativeVertical: p.Int64(3, "rel vertical"),
			IDType:           idType(p.Int64(4, "id type")),
			ID:               p.String(5, "id"),
			Track:            p.Int64(6, "track"),
			TurnRate:         p.Int64(7, "turn rate"),
			GroundSpeed:      p.Int64(8, "ground speed"),
			ClimbRate:        p.Float64(9, "climb rate"),
			AircraftType:     aircraftType(p.String(10, "aircraft type")),
		}, p.Err()
	})
}

// 1 = official ICAO 24-bit aircraft address
// 2 = stable FLARM ID (chosen by FLARM)
// 3 = anonymous ID, used if stealth mode is activated
func idType(v int64) string {
	switch v {
	case 1:
		return "official"
	case 2:
		return "flarm id"
	case 3:
		return "anonymous"
	}
	return "unknown"
}

// Hexadecimal value. Range: from 0 to F.
// Aircraft types:
// 0 = unknown
// 1 = glider / motor glider
// 2 = tow / tug plane
// 3 = helicopter / rotorcraft
// 4 = skydiver
// 5 = drop plane for skydivers
// 6 = hang glider (hard)
// 7 = paraglider (soft)
// 8 = aircraft with reciprocating engine(s)
// 9 = aircraft with jet/turboprop engine(s)
// A = unknown
// B = balloon
// C = airship
// D = unmanned aerial vehicle (UAV)
// E = unknown
// F = static object
func aircraftType(v string) string {
	switch v {
	case "1":
		return "glider / motor glider"
	case "2":
		return "tow / tug plane"
	case "3":
		return "helicopter / rotorcraft"
	case "4":
		return "skydiver"
	case "5":
		return "drop plane for skydivers"
	case "6":
		return "hang glider (hard)"
	case "7":
		return "paraglider (soft)"
	case "8":
		return "aircraft with reciprocating engine(s)"
	case "9":
		return "aircraft with jet/turboprop engine(s)"
	case "B":
		return "balloon"
	case "C":
		return "airship"
	case "D":
		return "unmanned aerial vehicle (UAV)"
	case "F":
		return "static object"
	}
	return "unknown"
}
