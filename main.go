package main

import (
	"suvi/src/bsky"
	"suvi/src/utils"
)

func main() {
	p := utils.CreatePost()
	bsky.Run(p)
}
