package flarmport

import "github.com/adrianmo/go-nmea"

// PGRMZ - Garmin's barometric altitude
// Syntax:
// Treat the following three versions as identical although FLARM currently only
// delivers the last one:
// PGRMZ,<Value>,F,3
// PGRMZ,<Value>,F
// PGRMZ,<Value>,F,2
type TypePGRMZ struct {
	nmea.BaseSentence `json:"-"`

	// Gives the barometric altitude in feet (1 ft = 0.3028 m) and can be negative.
	Altitude int64
}

func init() {
	nmea.MustRegisterParser("GRMZ", func(s nmea.BaseSentence) (nmea.Sentence, error) {
		// This example uses the package builtin parsing helpers
		// you can implement your own parsing logic also
		p := nmea.NewParser(s)
		return TypePGRMZ{
			BaseSentence: s,
			Altitude:     p.Int64(0, "alt"),
		}, p.Err()
	})
}
