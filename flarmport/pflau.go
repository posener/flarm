package flarmport

import "github.com/adrianmo/go-nmea"

// From spec:
// PFLAU: Operating status, priority intruder and obstacle warnings
// Syntax: PFLAU,<RX>,<TX>,<GPS>,<Power>,<AlarmLevel>,<RelativeBearing>,<AlarmType>,<RelativeVertical>,<RelativeDistance>,<ID>
type TypePFLAU struct {
	nmea.BaseSentence
	// Rx is the number of devices with unique IDs currently received regardless of the horizontal
	// or vertical separation
	// Decimal integer value. Range: from 0 to 99.
	Rx int64
	// Tx is the transmission status. 0 is OK, 1 is no transmission.
	Tx string
	// GPS status:
	//   0 = no GPS reception
	//   1 = 3d-fix on ground, i.e. not airborne
	//   2 = 3d-fix when airborne
	GPS string
	// Power level: 1 OK, 0 under or over voltage.
	Power int64
	// Alarm level:
	//
	// Decimal integer value. Range: from 0 to 3.
	//  0 = no alarm (also used for no-alarm traffic	information)
	// 	1 = alarm, 13-18 seconds to impact
	// 	2 = alarm, 9-12 seconds to impact
	// 	3 = alarm, 0-8 seconds to impact
	AlarmLevel int64
	// Decimal integer value. Range: -180 to 180.
	// Relative bearing in degrees from true ground track to
	// the intruder’s position. Positive values are clockwise. 0°
	// indicates that the object is exactly ahead. Field is empty
	// for non-directional targets or when no aircraft are
	// within range. For obstacle alarm and Alert Zone alarm,
	// this field is 0.
	RelativeBearing int64
	// Hexadecimal value. Range: from 0 to FF.
	// Type of alarm as assessed by FLARM
	// 0 = no aircraft within range or no-alarm traffic information
	// 2 = aircraft alarm
	// 3 = obstacle/Alert Zone alarm
	AlarmType string
	// Decimal integer value. Range: from -32768 to 32767.
	// Relative vertical separation in meters above own
	// position. Negative values indicate that the other aircraft
	// or obstacle is lower. Field is empty when no aircraft are
	// within range
	// For Alert Zone and obstacle warnings, this field is 0.
	RelativeVertical int64
	// Decimal integer value. Range: from 0 to 2147483647.
	// Relative horizontal distance in meters to the target or
	// obstacle. For non-directional targets this value is
	// estimated based on signal strength.
	// Field is empty when no aircraft are within range and no
	// alarms are generated.
	// For Alert Zone, this field is 0.
	RelativeDistance int64
	// The field is omitted for protocol version < 4.
	// 6-digit hexadecimal value (e.g. “5A77B1”) as
	// configured in the target’s PFLAC,,ID.
	// The interpretation is only delivered in <ID-Type> in the
	// PFLAA sentence (if received for the same aircraft).
	// The <ID> field is the ICAO 24-bit address for Mode-S
	// targets and a FLARM-generated ID for Mode-C targets.
	// The ID for Mode-C targets may change at any time.
	// Field is empty when no aircraft are within range and no
	// alarms are generated.
	// For obstacles this field is set to FFFFFF. In case of Alert
	// Zone warning, the FLARM ID of the Alert Zone station is
	// output.
	ID string
}

func init() {
	nmea.MustRegisterParser("FLAU", func(s nmea.BaseSentence) (nmea.Sentence, error) {
		// This example uses the package builtin parsing helpers
		// you can implement your own parsing logic also
		p := nmea.NewParser(s)
		ret := TypePFLAU{
			BaseSentence:     s,
			Rx:               p.Int64(0, "rx"),
			Tx:               tx(p.Int64(1, "tx")),
			GPS:              gpsStatus(p.Int64(2, "gps")),
			Power:            p.Int64(3, "power"),
			AlarmLevel:       p.Int64(4, "alarm level"),
			RelativeBearing:  p.Int64(5, "relative bearing"),
			AlarmType:        alarmType(p.String(6, "alarm type")),
			RelativeVertical: p.Int64(7, "relative vertical"),
			RelativeDistance: p.Int64(8, "relative distance"),
		}
		// Optional field, store error before parsing it.
		err := p.Err()
		ret.ID = p.String(9, "id")
		return ret, err
	})
}

// Tx is the transmission status. 0 is OK, 1 is no transmission.
func tx(v int64) string {
	switch v {
	case 0:
		return "OK"
	case 1:
		return "no transmission"
	}
	return "unknown"
}

//   0 = no GPS reception
//   1 = 3d-fix on ground, i.e. not airborne
//   2 = 3d-fix when airborne
func gpsStatus(v int64) string {
	switch v {
	case 0:
		return "no signal"
	case 1:
		return "valid on ground"
	case 2:
		return "valid airborne"
	}
	return "unknown"
}

// 0 = no aircraft within range or no-alarm traffic information
// 2 = aircraft alarm
// 3 = obstacle/Alert Zone alarm
func alarmType(v string) string {
	switch v {
	case "0":
		return "no alarm"
	case "2":
		return "aircraft"
	case "3":
		return "obstacle / zone"
	}
	return "unknown"
}
