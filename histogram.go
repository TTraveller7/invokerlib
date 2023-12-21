package invokerlib

type Histogram struct {
	name   string
	values []float64
}

func NewHistogram(name string, initialValues ...float64) *Histogram {
	vals := make([]float64, 0)
	if len(initialValues) > 0 {
		vals = initialValues
	}
	return &Histogram{
		name:   name,
		values: vals,
	}
}

func (h *Histogram) addValue(val float64) {
	h.values = append(h.values, val)
}
