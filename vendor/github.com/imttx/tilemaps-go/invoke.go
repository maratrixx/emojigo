package tilemaps

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"tilemaps-go/utils"
	"time"
)

var tilemapExt = ".png"

// 百度坐标转换墨卡托
func LatLng2MercatorPoint(p LngLatPoint) (point MercatorPoint) {
	var arr [10]float64

	var nlat float64

	if p.Latitude > 74 {
		nlat = 74
	} else {
		nlat = p.Latitude
	}

	if nlat < -74 {
		p.Latitude = -74
	} else {
		p.Latitude = nlat
	}

	array1Len := len(array1)
	for i := 0; i < array1Len; i++ {
		if p.Latitude >= float64(array1[i]) {
			arr = array2[i]
			break
		}
	}

	if len(arr) == 0 {
		for i := array1Len - 1; i >= 0; i-- {
			if p.Latitude <= float64(-array1[i]) {
				arr = array2[i]
				break
			}
		}
	}

	commPoint := Convertor(p.Longitude, p.Latitude, arr)

	return MercatorPoint{X: commPoint.X, Y: commPoint.Y}

}

// 墨卡托坐标转换百度坐标
func MercatorPoint2LatLng(p MercatorPoint) (point LngLatPoint) {
	var arr [10]float64
	p = MercatorPoint{X: math.Abs(p.X), Y: math.Abs(p.Y)}

	array3Len := len(array3)
	for i := 0; i < array3Len; i++ {
		if p.Y >= array3[i] {
			arr = array4[i]
			break
		}
	}

	commPoint := Convertor(p.X, p.Y, arr)

	return LngLatPoint{Longitude: commPoint.X, Latitude: commPoint.Y}
}

// 计算转换后的坐标点
func Convertor(x, y float64, params [10]float64) (point CommonPoint) {
	t := params[0] + params[1]*math.Abs(x)
	cC := math.Abs(y) / params[9]
	cF := params[2] + params[3]*cC + params[4]*cC*cC + params[5]*cC*cC*cC + params[6]*cC*cC*cC*cC + params[7]*cC*cC*cC*cC*cC + params[8]*cC*cC*cC*cC*cC*cC

	if x < 0 {
		t *= -1
	} else {
		t *= 1
	}

	if y < 0 {
		cF *= -1
	} else {
		cF *= 1
	}

	return CommonPoint{X: t, Y: cF}

}

// 计算瓦片坐标
func TilePosition(p MercatorPoint, level int) (point TilePoint) {
	x := int(math.Floor(p.X*math.Pow(2, float64(level-18))) / 256)
	y := int(math.Floor(p.Y*math.Pow(2, float64(level-18))) / 256)

	return TilePoint{X: x, Y: y, Level: level}
}

func SpliceTileMapPng(tiles []TilePoint, rect image.Rectangle) (img image.Image, err error) {

	co := image.NewNRGBA(rect)

	xPos, yPox, xMax := 0, 0, 0
	for _, tile := range tiles {
		if tile.X < xMax {
			xPos = 0
			yPox++
		}

		tileFilename := tileTargetFile(utils.GetStoragePath(), tile.Type, tile.Level, tile.X, tile.Y, tilemapExt)
		if !utils.CheckFileExists(tileFilename) {
			continue
		}

		tileFile, err := os.Open(tileFilename)
		if err != nil {
			continue
		}

		srcImg, _, err := image.Decode(tileFile)
		if err != nil {
			tileFile.Close()
			continue
		}

		tileRect := image.Rect(xPos*256, yPox*256, xPos*256+256, yPox*256+256)
		draw.Draw(co, tileRect, srcImg, image.ZP, draw.Src)

		xPos++
		xMax = tile.X

		tileFile.Close()
	}

	return co, nil
}

// 根据瓦片坐标计算区域大小
func GetNRGBARect(src, dst TilePoint) (rect image.Rectangle) {
	xMax := int(math.Abs(float64(src.X-dst.X)) + 1)
	yMax := int(math.Abs(float64(src.Y-dst.Y)) + 1)

	rect = image.Rect(0, 0, xMax*256, yMax*256)
	return
}

// 随机获取一个domain
func randomDomain() (domain string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return apiDomains[r.Intn(len(apiDomains))]
}

func BeginMakeTilemap(tm Tilemap) (image.Image, error) {
	// 构建瓦片地图
	return makeTilemapHandler(tm)
}

func makeTilemapHandler(tm Tilemap) (img image.Image, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("[ERROR] make tilemap panic: %#v", e)
			err = fmt.Errorf("%#v", e)
		}
	}()

	err = CheckTilemapParams(&tm)
	if err != nil {
		return
	}

	// 地图坐标
	s1 := LngLatPoint{Longitude: tm.Startx, Latitude: tm.Starty}
	s2 := LngLatPoint{Longitude: tm.Endx, Latitude: tm.Endy}

	// 瓦片坐标
	t1 := TilePosition(LatLng2MercatorPoint(s1), tm.Level)
	t2 := TilePosition(LatLng2MercatorPoint(s2), tm.Level)

	var tiles []TilePoint

	// 截取domain
	url := randomDomain()

	// 建筑地图类型
	styles := MapTypes[tm.Type]

	react := GetNRGBARect(t1, t2)

	// 多核并行处理，来充分利用机器CPU
	runtime.GOMAXPROCS(runtime.NumCPU())

	ymin, ymax, xmin, xmax := t1.Y, t2.Y, t1.X, t2.X

	if ymin > ymax {
		ymin, ymax = ymax, ymin
	}

	if xmin > xmax {
		xmin, xmax = xmax, xmin
	}

	storagePath := utils.GetStoragePath()
	for j := ymax; j >= ymin; j-- {
		wait := sync.WaitGroup{}
		wait.Add(xmax - xmin + 1) //按照行进行wait等待，防止goroutine开启太多

		for i := xmin; i <= xmax; i++ {
			targetFile := tileTargetFile(storagePath, tm.Type, tm.Level, i, j, tilemapExt)
			utils.CheckDirAndMkdir(filepath.Dir(targetFile))
			go utils.SaveImageAsync(fmt.Sprintf(url, i, j, tm.Level)+styles, targetFile, &wait)
			tiles = append(tiles, TilePoint{X: i, Y: j, Level: tm.Level, Type: tm.Type})
		}

		wait.Wait()
	}

	img, err = SpliceTileMapPng(tiles, react)

	return
}

func tileTargetFile(basedir string, tiletype, level, x, y int, ext string) (filename string) {
	return filepath.Join(basedir, strconv.Itoa(tiletype), strconv.Itoa(level), strconv.Itoa(x), fmt.Sprintf("%d%s", y, ext))
}

func makeOutputTempFile(baseDir, ext, prefix string) (filename string) {
	timeStr := time.Now().Format("20060102150405")
	filename = filepath.Join(baseDir, fmt.Sprintf("%s%s%s", prefix, timeStr, ext))
	return
}

// 参数检查
func CheckTilemapParams(tm *Tilemap) error {

	if !utils.SliceKeyExists(tm.Type, MapTypes) {
		return errors.New("type参数不合法")
	}

	if tm.Level < LevelMin || tm.Level > LevelMax {
		return errors.New("level参数不合法")
	}

	// 地图坐标
	s1 := LngLatPoint{Longitude: tm.Startx, Latitude: tm.Starty}
	s2 := LngLatPoint{Longitude: tm.Endx, Latitude: tm.Endy}

	// 瓦片坐标
	t1 := TilePosition(LatLng2MercatorPoint(s1), tm.Level)
	t2 := TilePosition(LatLng2MercatorPoint(s2), tm.Level)

	if t1.X == t2.X || t1.Y == t2.Y {
		return errors.New("坐标范围不合法，请重新输入")
	}

	if t1.X <= 0 || t1.Y <= 0 || t2.X <= 0 || t2.Y <= 0 {
		return errors.New("坐标参数不合法，目前只支持国内范围")
	}

	if math.Abs(float64(t1.X-t2.X)*float64(t1.Y-t2.Y)) > TileNumMax {
		return errors.New("坐标范围过大，请分成多块生成")
	}

	return nil
}
