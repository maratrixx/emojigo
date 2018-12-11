package tilemaps

// 经纬度坐标
type LngLatPoint struct {
	Longitude, Latitude float64
}

// 墨卡托坐标
type MercatorPoint struct {
	X, Y float64
}

// 通用坐标
type CommonPoint MercatorPoint

type TilePoint struct {
	X, Y, Level, Type int
}

type Tilemap struct {
	Type   int
	Startx float64
	Starty float64
	Endx   float64
	Endy   float64
	Level  int
}
