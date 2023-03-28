package speeding

type Camera struct {
	Road  uint16
	Mile  uint16
	Limit uint16
}

type Observation struct {
	timestamp  uint32
	mileMarker uint16
}

func (o *Observation) IsBefore(other *Observation) bool {
	return o.timestamp < other.timestamp
}

func (o *Observation) SpeedBetween(other *Observation) float64 {
	timeDiff := (float64(o.timestamp) - float64(other.timestamp)) / 3600.0
	if timeDiff < 0 {
		timeDiff *= -1
	}

	distance := float64(o.mileMarker) - float64(other.mileMarker)
	if distance < 0 {
		distance *= -1
	}

	return distance / timeDiff
}
