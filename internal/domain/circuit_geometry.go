package domain

type CircuitGeometry struct {
	SessionKey       int          `json:"session_key"`
	Points           []TrackPoint `json:"points"`
	EstimatedWidthM  float64      `json:"estimated_width_m"`
	Attribution      string       `json:"attribution"`
	SourceURL        string       `json:"source_url"`
	GeometryAccuracy string       `json:"geometry_accuracy"`
}
