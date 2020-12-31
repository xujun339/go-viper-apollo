package viper_helper

import (
	"fmt"
	"time"
)

type Logg struct {

}

func (l Logg) Debug(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

func (l Logg) Info(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

func (l Logg) Warn(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

func (l Logg) Error(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

