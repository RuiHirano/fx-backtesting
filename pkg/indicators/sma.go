package indicators

// SMA represents a Simple Moving Average indicator
type SMA struct {
	period int
	values []float64
	sum    float64
	ready  bool
}

// NewSMA creates a new Simple Moving Average with the specified period
func NewSMA(period int) *SMA {
	return &SMA{
		period: period,
		values: make([]float64, 0, period),
		sum:    0.0,
		ready:  false,
	}
}

// Update adds a new value and returns the current SMA value
func (s *SMA) Update(value float64) float64 {
	if len(s.values) < s.period {
		// Still filling up the initial period
		s.values = append(s.values, value)
		s.sum += value
		
		if len(s.values) == s.period {
			s.ready = true
		}
		
		// Return average of values so far
		return s.sum / float64(len(s.values))
	} else {
		// Sliding window: remove oldest, add newest
		oldest := s.values[0]
		s.sum = s.sum - oldest + value
		
		// Shift values and add new one
		copy(s.values, s.values[1:])
		s.values[s.period-1] = value
		
		return s.sum / float64(s.period)
	}
}

// GetValue returns the current SMA value
func (s *SMA) GetValue() float64 {
	if len(s.values) == 0 {
		return 0.0
	}
	return s.sum / float64(len(s.values))
}

// IsReady returns true if the indicator has enough data points
func (s *SMA) IsReady() bool {
	return s.ready
}

// Reset clears all data and resets the indicator
func (s *SMA) Reset() {
	s.values = s.values[:0]
	s.sum = 0.0
	s.ready = false
}

// GetPeriod returns the period of the SMA
func (s *SMA) GetPeriod() int {
	return s.period
}