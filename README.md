# emojigo
自动化生成微信文字斗图表情包

# Todo
* [x] 表情主图
* [x] 表情缩略图
* [x] 详情页横幅
* [x] 表情封面图
* [x] 聊天面板图标

## Requirements

NULL

## Installation

```
go get github.com/imttx/emojigo
```

## Documentation

TODO

## Example
``` go

	m := &service.MainImg{
		TextSlice: []string{
			"以梦为马，越骑越傻",
			"诗和远方，越远越脏",
			"执子之手，如同猪肘",
			"故事和酒，淘宝都有",
			"春风十里，吹不死你",
			"心有猛虎，像二百五",
			"嘘寒问暖，不如巨款",
			"闲庭信步，忘穿秋裤",
			"面朝大海，笑出精彩",
			"白云苍狗，你比我丑",
			"岁月静好，全靠胆小",
			"你若安好，打支付宝",
			"寒风十里，冷的飞起",
			"没有梦想，过得特爽",
			"不愿将就，装逼没够",
			"随遇而安，得脑血栓",
		},
		Path:         "/tmp/emoji",
		FontSize:     100,
		IconTitle:    "搞笑，文字",
		IconColor:    color.Transparent,
		ProfileTitle: "岁月静好全靠胆小",
		ProfileColor: color.NRGBA{239, 233, 87, 255},
		Theme:        "",
	}

	err := m.Do()
  
  fmt.Println(err)

```
