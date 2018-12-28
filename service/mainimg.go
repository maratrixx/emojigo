// 表情包主图绘制
// 素材要求参考：https://sticker.weixin.qq.com

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
	fontSize   float64 = 100
	fontDpi    float64 = 72
	maxNum             = 24
	minNum             = 16
	sep                = "|"
)

type MainImg struct {
	Theme        string         // 表情包主题
	TextSlice    []string       // 文字列表
	IconTitle    string         // 图标文案
	IconColor    color.Color    // 图标底色
	ProfileTitle string         // 详情页横幅
	ProfileColor color.Color    // 详情页底色
	Path         string         // 保存路径
	FontFile     string         // 字体文件
	FontSize     float64        // 字体大小
	parsedFont   *truetype.Font // 解析后的字体引用
}

// init 初始化
func init() {
	rootDir, err := com.GetSrcPath("emojigo")
	if err != nil {
		panic(err)
	}

	parsedFont, err = parseFont(filepath.Join(rootDir, "public/fonts", "RuanMengTi-2.ttf"))
	if err != nil {
		panic(err)
	}
}

// parseFont 解析字体文件
func parseFont(file string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return freetype.ParseFont(fontBytes)
}

// SetFont 设置字体
func (m *MainImg) SetFont(fontFile string) error {
	m.parsedFont = parsedFont

	if !com.IsFile(fontFile) {
		return fmt.Errorf("file: %s is illegal", fontFile)
	}

	ft, err := parseFont(fontFile)
	if err != nil {
		return err
	}

	m.parsedFont = ft
	return nil
}

// SetSep 设置换行标识符
func (m *MainImg) SetSep(s string) {
	sep = s
}

// Do 执行绘图
func (m *MainImg) Do() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// 检查数量
	if len(m.TextSlice) != minNum && len(m.TextSlice) != maxNum {
		return errors.New("text nums is illegal")
	}

	// 设置字体
	m.SetFont(m.FontFile)

	// 设置字体大小
	if m.FontSize <= 0 {
		m.FontSize = fontSize
	}

	wg := &sync.WaitGroup{}
	for k, v := range m.TextSlice {
		wg.Add(2)
		k++
		// 主图
		go makeGif(wg, v, fmt.Sprintf("%s/%02d.gif", m.Path, k), image.Rect(0, 0, 240, 240), m.FontSize, color.Transparent, m.parsedFont)
		// 表情缩略图
		go makePng(wg, v, fmt.Sprintf("%s/%02d.png", m.Path, k), image.Rect(0, 0, 120, 120), m.FontSize, color.White, m.parsedFont)
	}

	wg.Add(3) // 生成图标&&封面图&&横幅图

	// 聊天页图标
	go makePng(wg, m.IconTitle, fmt.Sprintf("%s/icon.png", m.Path), image.Rect(0, 0, 50, 50), m.FontSize, m.IconColor, m.parsedFont)

	// 表情封面
	go makePng(wg, m.IconTitle, fmt.Sprintf("%s/cover.png", m.Path), image.Rect(0, 0, 240, 240), m.FontSize, m.IconColor, m.parsedFont)

	// 详情页横幅
	go makePng(wg, m.ProfileTitle, fmt.Sprintf("%s/header.png", m.Path), image.Rect(0, 0, 750, 400), m.FontSize, m.ProfileColor, m.parsedFont)

	// 等待完成
	wg.Wait()

	return nil
}

func makeGif(wg *sync.WaitGroup, text string, filename string, rect image.Rectangle, fs float64, co color.Color, ft *truetype.Font) {
	defer wg.Done()

	img := image.NewPaletted(rect, []color.Color{co, color.Black})

	err := drawText(img, text, color.Black, fs, fontDpi, true, ft)
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

func makePng(wg *sync.WaitGroup, text string, filename string, rect image.Rectangle, fs float64, co color.Color, ft *truetype.Font) {
	defer wg.Done()
	img := image.NewNRGBA(rect)

	// 背景色设置
	draw.Draw(img, img.Bounds(), image.NewUniform(co), image.ZP, draw.Src)

	// 绘制文字
	err := drawText(img, text, color.Black, fs, fontDpi, true, ft)
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
func drawText(img draw.Image, text string, co color.Color, fs float64, dpi float64, split bool, ft *truetype.Font) (err error) {
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
	textSlice := []string{text}
	if split {
		textSlice = strings.Split(text, sep)
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
