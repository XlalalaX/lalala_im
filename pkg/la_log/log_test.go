package log

import "testing"

func TestNewLog(t *testing.T) {
	InitLog("D:\\Goland\\go_xm\\lalala_im\\log\\la_log.log", "INFO", "json")
	Panic("lalala", Log)
}
