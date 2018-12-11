package main

import (
	"emojigo/service"
)

func main() {
	doMainImg()
}

func doMainImg() {
	m := &service.MainImg{
		Width:  240,
		Height: 240,
		Num:    16,
		TextSlice: []string{
			"你摊上事了",
			"人丑还特矫情",
			"你们先聊,我自闭一会",
			"可爱的我又出现了",
			"此时场面,略显尴尬",
			"凭实力尬聊"},
		Path: "/Users/yafeng5/.gvm/pkgsets/go1.11/global/src/emojigo/public",
	}

	m.Do()
}
