package main

import "infro.io/infro-core/cmd/comment"

func main() {
	if err := comment.NewDiffsCommand().Execute(); err != nil {
		panic(err)
	}
}
