# tilemaps-go

基于 Golang 协程开发实现的瓦片地图下载服务

## Requirements

NULL

## Installation

```
go get github.com/Liyafeng/tilemaps-go
```

## Documentation

TODO

## Example
``` go

tm := Tilemap{
	Type:   2,
	Level:  16,
	Startx: 116.397558,
	Starty: 39.930505,
	Endx:   116.412255,
	Endy:   39.913462,
}

img, _ := tilemaps.BeginMakeTilemap(tm)

imgFile, _ := os.Create("./test.png")

png.Encode(imgFile, img)

```
