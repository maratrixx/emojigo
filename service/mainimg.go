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
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/Unknwon/com"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"github.com/golang/freetype"
	_ "github.com/lifei6671/gocaptcha"
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
		//makePng(v, fmt.Sprintf("%s/%02d.png", m.Path, k), rect)
	}
	return nil
}

func makeGif(text string, filename string, rect image.Rectangle) {
	img := image.NewPaletted(rect, []color.Color{color.Transparent, color.Black})
	err := drawText(img, text, image.Point{50, 80}, color.Black, 40, 80)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(500)
	}

	fh, _ := os.Create(filename)
	defer fh.Close()

	g := &gif.GIF{Image: []*image.Paletted{img}, Delay: []int{30}, LoopCount: 0}
	gif.EncodeAll(fh, g)

	os.Exit(201)
}

func makePng(text string, filename string, rect image.Rectangle) {
	img := image.NewNRGBA(rect)
	err := drawText(img, text, image.Point{50, 80}, color.Black, 40, 80)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(500)
	}

	fh, _ := os.Create(filename)
	defer fh.Close()

	png.Encode(fh, img)
	os.Exit(202)
}

// DrawText 绘制文字
func drawText(img draw.Image, text string, p image.Point, co color.Color, fs float64, dpi float64) (err error) {

	//读取字体数据
	rootDir, _ := com.GetSrcPath("emojigo")
	fontBytes, err := ioutil.ReadFile(filepath.Join(rootDir, "public/fonts", "AaSuiBianYiDian-2.ttf"))

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

	f.SetClip(img.Bounds())

	//设置输出的图片
	f.SetDst(img)

	//设置字体颜色
	f.SetSrc(image.NewUniform(co))

	//设置尺寸
	fs, err = autoFontSize(img, []string{text}, fs, f)
	if err != nil {
		return err
	}
	f.SetFontSize(fs)

	//调整位置
	fontHeight := f.PointToFixed(fs).Ceil() * 3 / 4
	p.Y = img.Bounds().Dy() - (img.Bounds().Dy()-fontHeight)/2
	p.X = (img.Bounds().Dx() - f.PointToFixed(fs).Round()*utf8.RuneCountInString(text)) / 2
	pt := freetype.Pt(p.X, p.Y)

	//绘制文字
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

// autoFontSize 自动我微调字体大小来适应整个画布
func autoFontSize(img image.Image, textList []string, fontSize float64, ft *freetype.Context) (float64, error) {
	maxTextLen := 0
	for _, text := range textList {
		l := utf8.RuneCountInString(text)
		if l > maxTextLen {
			maxTextLen = l
		}
	}

	for {
		if fontSize <= 0 {
			return fontSize, errors.New("fontsize is negative")
		}

		w := ft.PointToFixed(fontSize).Round() * maxTextLen
		if w > img.Bounds().Dx() {
			fontSize--
			continue
		}

		h := ft.PointToFixed(fontSize).Ceil() * len(textList)
		if h > img.Bounds().Dy() {
			fontSize--
			continue
		}

		break
	}

	return fontSize, nil
}
