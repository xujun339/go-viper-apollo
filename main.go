package main

import (
	"fmt"
	"time"

	"github.com/gin-sasuke/sasuke/pkg/apollo"
	"github.com/gin-sasuke/sasuke/pkg/viper_helper"
)

func main() {
	//var apolloUrl string
	//flag.StringVar(&apolloUrl, "apollo", "", "apllo地址")
	//flag.Parse()
	//viper_helper.InitApolloUrl(apolloUrl)
	/*
		//InitApollo()
		//BuildConfig("aa","yml")
		//BuildConfig("bb","json")
		//GetApolloConfig()
		//InitWatcher()
		//
	*/

	apollo.InitViperConfig("config", "properties", "config")
	srv := apollo.New(viper_helper.Logg{})
	srv.AddYMLConfig("application")
	srv.AddConfig("db")
	if initConfigErr := viper_helper.StartApollo(srv); initConfigErr != nil {
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

	select {}
}
