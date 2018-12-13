package main

import (
	"emojigo/service"
	"fmt"

	"github.com/Unknwon/com"
)

func main() {
	doMainImg()
}

func doMainImg() {
	tmpDir, _ := com.GetSrcPath("emojigo/public")
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
			"不忘初心，放得始终",
			"寒风十里，冷的飞起",
			"没有梦想，过得特爽",
			"不愿将就，装逼没够",
			"随遇而安，得脑血栓",
			"鱼与熊掌，逼良为娼",
		},
		Path:  tmpDir,
		Title: "毁人不倦",
	}

	err := m.Do()

	fmt.Println(err)
}
