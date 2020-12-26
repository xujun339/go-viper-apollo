package main

import (
	"flag"
	"fmt"
	"github.com/gin-sasuke/sasuke/pkg/viper_helper"
	"time"
)

func main() {
	var apolloUrl string
	flag.StringVar(&apolloUrl, "apollo", "", "apllo地址")
	flag.Parse()
	viper_helper.InitApolloUrl(apolloUrl)

	if initConfigErr := viper_helper.InitLocalConfig("config"); initConfigErr != nil {
		fmt.Println(initConfigErr)
	} else {
		go func() {
			timer1 := time.NewTicker(1 * time.Second)
			for {
				<-timer1.C
				fmt.Printf("%+v\n", viper_helper.Configmap["logging-level"].GetViper().AllSettings())
				fmt.Printf("%+v\n", viper_helper.Configmap["application"].GetViper().AllSettings())
			}
		}()
	}


	select {

	}

}
