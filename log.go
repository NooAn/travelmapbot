package main

import (
	"fmt"
)

var debug = true
var tgbot = &TBWrap{}

func IsDebug() bool {
	return debug
}
func EnableDebug() {
	Log("Debug mode started")
	debug = true
}
func DisableDebug() {
	Log("Debug mode stop")
	debug = false
}
func LogPanic(v ...interface{}){
	s := fmt.Sprintln(v...)
	Log(s)
}
func Logf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	Log(s)
}
func Log(message string) {
		if (debug == true) {
			fmt.Println(message)
		}
		if (BOT_ACTIVE_LOGER == 1) {
			tgbot.Send(GROUP_ID_LOGER, message)
		}
}

func CheckErr(e error, message string) {
	if e != nil {
		Log("Error: " + e.Error() + " \nFailed text: " + message)
	}
}