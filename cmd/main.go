package main

import "infro.io/infro-core/cmd/root"

func main() {
	if err := root.NewCommand().Execute(); err != nil {
		panic(err)
	}
}
