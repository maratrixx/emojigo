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
			"人丑还特矫情",
			"我能撩你吗",
			"你摊上事了",
			"你们先聊，我自闭一会",
			"可爱的我，又出现了",
			"此时场面，略显尴尬",
			"凭实力尬聊"},
		Path:  tmpDir,
		Title: "实力尬聊",
	}

	err := m.Do()

	fmt.Println(err)
}
