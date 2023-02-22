package la_email

import "testing"

func TestSendMail(t *testing.T) {
	//InitEmail("1633159337@qq.com", "啦啦啦", "ebuwiuvkxfyfeffc", "smtp.qq.com", 465)
	InitEmail("xu1633159337@163.com", "lalala", "PNTWLIKHBRRUBMSS", "smtp.163.com", 465)
	err := SendMail([]string{"1633159337@qq.com"}, "lalala", "lalal2")
	if err != nil {
		panic(err)
	}
}
