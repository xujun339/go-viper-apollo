package main

import (
	"fmt"
	"github.com/gin-sasuke/sasuke/pkg/viper_helper"
	"time"
)

func main() {

	viper_helper.InitLocalConfig("config")
	go func() {
		timer1 := time.NewTicker(1 * time.Second)
		for {
			<-timer1.C
			fmt.Printf("%+v\n", viper_helper.Configmap["logging-level"].GetViper().AllSettings())
			fmt.Printf("%+v\n", viper_helper.Configmap["application"].GetViper().AllSettings())
		}
	}()

	select {

	}

}
