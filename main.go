package main

import (
	"YTSubs/base"
	"fmt"
)

func main() {
	fmt.Println(base.Extract_channel_id("https://www.youtube.com/@pewdiepie"))
	fmt.Println(base.Extract_channel_id("https://www.youtube.com/channel/UCyseFvMP4mZVlU5iEEbAamA"))
}
