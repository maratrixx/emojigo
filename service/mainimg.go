// 表情包主图绘制

/**
素材要求:
- GIF 格式，240 ×240 像素，真人拍摄表情、截屏表情每张不超过500 KB，卡通表情及其它每张不超过 100 KB
- 数量只能是 16 或 24 张
- 同一套表情主图须全部是动态或全部是静态
- 动态表情设置循环播放
- 同一套表情中各表情风格须统一
*/

package service

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Unknwon/com"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"github.com/golang/freetype"
	_ "github.com/imttx/tilemaps-go"
	"golang.org/x/image/math/fixed"
)

const (
	MaxWidth  = 240
	MadHeight = 240
	MinNum    = 16
	MaxNum    = 24
)

type MainImg struct {
	Width     int
	Height    int
	Num       int
	TextSlice []string
	Path      string
}

func (m *MainImg) Do() error {
	rect := image.Rect(0, 0, m.Width, m.Height)
	for k, v := range m.TextSlice {
		makeGif(v, fmt.Sprintf("%s/%02d.gif", m.Path, k), rect)
	}
	return nil
}

func (m *MainImg) DoTest() error {
	f1, err := os.Create(m.Path + "/test.gif")
	if err != nil {
		fmt.Println(err)
	}
	defer f1.Close()

	p1 := image.NewPaletted(image.Rect(0, 0, 110, 110), palette.Plan9)
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			p1.Set(50, y, color.RGBA{uint8(x), uint8(y), 255, 255})
		}
	}
	p2 := image.NewPaletted(image.Rect(0, 0, 210, 210), palette.Plan9)
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			p2.Set(x, 50, color.RGBA{uint8(x * x % 255), uint8(y * y % 255), 0, 255})
		}
	}
	g1 := &gif.GIF{
		Image:     []*image.Paletted{p1, p2},
		Delay:     []int{30, 30},
		LoopCount: 0,
	}
	gif.EncodeAll(f1, g1) //保存到文件中

	return nil
}

func makeGif(text string, filename string, rect image.Rectangle) {
	img := image.NewPaletted(rect, []color.Color{color.White, color.Black})
	err := DrawText(img, text, image.Point{20, 20}, color.Black, 20, 80, true)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(500)
	}

	fh, _ := os.Create(filename)
	defer fh.Close()

	g := &gif.GIF{Image: []*image.Paletted{img}, Delay: []int{30}, LoopCount: -1}
	gif.EncodeAll(fh, g)

	os.Exit(200)
}

// DrawText 绘制文字
func DrawText(img draw.Image, text string, p image.Point, co color.Color, fs float64, dpi float64, center bool) (err error) {

	//读取字体数据
	rootDir, _ := com.GetSrcPath("emojigo")
	fontBytes, err := ioutil.ReadFile(filepath.Join(rootDir, "public/fonts", "方正兰亭纤黑简体.ttf"))

	if err != nil {
		return err
	}

	//载入字体数据
	ft, err := freetype.ParseFont(fontBytes)

	f := freetype.NewContext()

	//设置分辨率
	f.SetDPI(dpi)

	//设置字体
	f.SetFont(ft)

	//设置尺寸
	f.SetFontSize(fs)

	f.SetClip(img.Bounds())

	//设置输出的图片
	f.SetDst(img)

	//设置字体颜色
	f.SetSrc(image.NewUniform(co))

	//根据是否居中设置字体的位置
	var pt fixed.Point26_6

	//widthTotal := measure(dpi, fs, text, ft)
	//p.X = (img.Bounds().Dx() - widthTotal) / 2
	//p.Y = (img.Bounds().Dy() - (widthTotal / utf8.RuneCountInString(text))) / 2

	if center {
		// 居中，需要计算调整X坐标点
		p.X = (img.Bounds().Dx() - measure(dpi, fs, text, ft)) / 2
		pt = freetype.Pt(p.X, p.Y)
	} else {
		// 非居中
		pt = freetype.Pt(p.X, p.Y)

	}

	// 绘制文字
	_, err = f.DrawString(text, pt)
	return
}

// measure 用于计算文本宽度
func measure(dpi, size float64, txt string, fnt *truetype.Font) int {
	opt := &truetype.Options{
		DPI:  dpi,
		Size: size,
	}
	face := truetype.NewFace(fnt, opt)

	return font.MeasureString(face, txt).Round()
}
