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
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/Unknwon/com"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"github.com/golang/freetype"
	_ "github.com/lifei6671/gocaptcha"
)

var (
	parsedFont *truetype.Font
)

type MainImg struct {
	Width     int
	Height    int
	Num       int
	TextSlice []string
	Path      string
	Title     string
	FontFile  string
}

func init() {
	rootDir, err := com.GetSrcPath("emojigo")
	if err != nil {
		panic(err)
	}

	fontBytes, _ := ioutil.ReadFile(filepath.Join(rootDir, "public/fonts", "RuanMengTi-2.ttf"))
	parsedFont, err = truetype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}
}

func (m *MainImg) SetFont(fontFile string) {

}

func (m *MainImg) Do() error {
	wg := &sync.WaitGroup{}
	for k, v := range m.TextSlice {
		wg.Add(2)
		go makeGif(wg, v, fmt.Sprintf("%s/%02d.gif", m.Path, k), image.Rect(0, 0, 240, 240), 50)
		go makePng(wg, v, fmt.Sprintf("%s/%02d.png", m.Path, k), image.Rect(0, 0, 240, 240), 50)
	}

	wg.Add(2) // 生成图标&&封面图
	go makePng(wg, m.Title, fmt.Sprintf("%s/icon.png", m.Path), image.Rect(0, 0, 50, 50), 50)
	go makePng(wg, m.Title, fmt.Sprintf("%s/cover.png", m.Path), image.Rect(0, 0, 240, 240), 100)

	wg.Wait()
	return nil
}

func makeGif(wg *sync.WaitGroup, text string, filename string, rect image.Rectangle, fs float64) {
	defer wg.Done()
	img := image.NewPaletted(rect, []color.Color{color.Transparent, color.Black})
	err := drawText(img, text, color.Black, fs, 72, true)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fh, _ := os.Create(filename)
	defer fh.Close()

	g := &gif.GIF{Image: []*image.Paletted{img}, Delay: []int{30}, LoopCount: 0}
	gif.EncodeAll(fh, g)

	fmt.Println(text, "gif done")
}

func makePng(wg *sync.WaitGroup, text string, filename string, rect image.Rectangle, fs float64) {
	defer wg.Done()
	img := image.NewNRGBA(rect)
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	err := drawText(img, text, color.Black, fs, 72, true)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fh, _ := os.Create(filename)
	defer fh.Close()

	png.Encode(fh, img)
	fmt.Println(text, "png done")
}

// DrawText 绘制文字
func drawText(img draw.Image, text string, co color.Color, fs float64, dpi float64, split bool) (err error) {
	//读取字体数据
	rootDir, _ := com.GetSrcPath("emojigo")
	fontBytes, err := ioutil.ReadFile(filepath.Join(rootDir, "public/fonts", "RuanMengTi-2.ttf"))

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

	// TODO 了解一下
	//f.SetHinting(50)

	//设置尺寸
	textSlice := strings.Split(text, "，")
	fs, err = autoFontSize(img, textSlice, fs, f)
	if err != nil {
		return err
	}
	f.SetFontSize(fs)

	//调整位置
	fontHeight := f.PointToFixed(fs).Ceil() * 3 / 4
	pY := (img.Bounds().Dy()-fontHeight*len(textSlice))/2 + fontHeight
	for _, v := range textSlice {
		pX := (img.Bounds().Dx() - f.PointToFixed(fs).Round()*utf8.RuneCountInString(v)) / 2
		pt := freetype.Pt(pX, pY)
		//绘制文字
		_, err = f.DrawString(v, pt)

		pY += fontHeight
	}

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
