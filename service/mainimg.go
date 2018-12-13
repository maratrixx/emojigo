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
)

var (
	parsedFont *truetype.Font
	fontSize   float64 = 50
	fontDpi    float64 = 72
	maxNum             = 24
	minNum             = 16
)

type MainImg struct {
	TextSlice []string
	Path      string
	Title     string
	FontFile  string
}

// init 初始化
func init() {
	rootDir, err := com.GetSrcPath("emojigo")
	if err != nil {
		panic(err)
	}
	parseFont(filepath.Join(rootDir, "public/fonts", "RuanMengTi-2.ttf"))
}

// parseFont 解析字体文件
func parseFont(file string) error {
	fontBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	parsedFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	return nil
}

// SetFont 设置字体
func (m *MainImg) SetFont(fontFile string) error {
	if !com.IsFile(fontFile) {
		return fmt.Errorf("file: %s is illegal", fontFile)
	}
	return parseFont(fontFile)
}

// Do 执行绘图
func (m *MainImg) Do() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// 检查数量
	//if len(m.TextSlice) != minNum && len(m.TextSlice) != maxNum {
	//	return errors.New("text nums is illegal")
	//}

	// 设置字体
	m.SetFont(m.FontFile)

	wg := &sync.WaitGroup{}
	for k, v := range m.TextSlice {
		wg.Add(2)
		k++
		go makeGif(wg, v, fmt.Sprintf("%s/%02d.gif", m.Path, k), image.Rect(0, 0, 240, 240), 50, true)
		go makePng(wg, v, fmt.Sprintf("%s/%02d.png", m.Path, k), image.Rect(0, 0, 240, 240), 50, false)
	}

	wg.Add(3) // 生成图标&&封面图&&横幅图
	go makePng(wg, m.Title, fmt.Sprintf("%s/icon.png", m.Path), image.Rect(0, 0, 50, 50), 50, true)
	go makePng(wg, m.Title, fmt.Sprintf("%s/cover.png", m.Path), image.Rect(0, 0, 240, 240), 100, true)
	go makePng(wg, m.Title, fmt.Sprintf("%s/header.png", m.Path), image.Rect(0, 0, 750, 400), 100, false)

	// 等待完成
	wg.Wait()

	return nil
}

func makeGif(wg *sync.WaitGroup, text string, filename string, rect image.Rectangle, fs float64, transparent bool) {
	defer wg.Done()

	// 透明度设置
	var colors []color.Color
	if transparent {
		colors = append(colors, color.Transparent, color.Black)
	} else {
		colors = append(colors, color.White, color.Black)
	}

	img := image.NewPaletted(rect, colors)

	err := drawText(img, text, color.Black, fs, fontDpi, true)
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

func makePng(wg *sync.WaitGroup, text string, filename string, rect image.Rectangle, fs float64, transparent bool) {
	defer wg.Done()
	img := image.NewNRGBA(rect)

	// 透明度设置
	if transparent {
		draw.Draw(img, img.Bounds(), image.Transparent, image.ZP, draw.Src)
	} else {
		draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	}

	err := drawText(img, text, color.Black, fs, fontDpi, true)
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
	f := freetype.NewContext()

	//设置分辨率
	f.SetDPI(dpi)

	//设置字体
	f.SetFont(parsedFont)

	f.SetClip(img.Bounds())

	//设置输出的图片
	f.SetDst(img)

	//设置字体颜色
	f.SetSrc(image.NewUniform(co))

	// TODO 了解一下
	//f.SetHinting(50)

	//设置尺寸
	textSlice := []string{text}
	if split {
		textSlice = strings.Split(text, "，")
	}
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

		pY += fontHeight * 4 / 3
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
