package user

type resolution struct {
	w int
	h int
}

var resolutions = map[resolution]struct{}{
	resolution{h: 80, w: 60}:   {},
	resolution{h: 90, w: 90}:   {},
	resolution{h: 100, w: 100}: {},
	resolution{h: 120, w: 120}: {},
	resolution{h: 150, w: 150}: {},
	resolution{h: 175, w: 175}: {},
	resolution{h: 200, w: 200}: {},
	resolution{h: 220, w: 220}: {},
	resolution{h: 250, w: 250}: {},
}
