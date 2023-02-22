package conf

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
	"lalala_im/pkg/la_auth"
	"lalala_im/pkg/la_email"
	log "lalala_im/pkg/la_log"
	"lalala_im/servers/receive_server/internal/db"
)

type (
	Config struct {
		App        `yaml:"app"`
		Data       ` yaml:"data"`
		Server     ` yaml:"server"`
		Log        `yaml:"log"`
		ThirdParty `yaml:"third_party"`
	}
	App struct {
		Auth    `yaml:"auth"`
		Name    string `yaml:"name" env:"APP_NAME"`
		Host    string `yaml:"host" env:"APP_HOST"`
		RunMode string `yaml:"run_mode" env:"RUN_MODE"`
	}
	Data struct {
		Mongo `yaml:"mongo"`
		Redis `yaml:"redis"`
	}
	Server struct {
		Http `yaml:"http"`
	}
	ThirdParty struct {
		Email  `yaml:"email"`
		AliOss `yaml:"ali_oss"`
	}
	Http struct {
		Port         int `yaml:"port" env:"SERVER_HTTP_PORT"`
		ReadTimeout  int `yaml:"read_timeout" env:"SERVER_HTTP_READ_TIMEOUT"`
		WriteTimeout int `yaml:"write_timeout" env:"SERVER_HTTP_WRITE_TIMEOUT"`
	}

	Mongo struct {
		Url      string `yaml:"url" env:"DATA_MONGO_URL"`
		Database string `yaml:"database" env:"DATA_MONGO_DATABASE"`
	}
	Redis struct {
		ADDR     string `yaml:"addr" env:"DATA_REDIS_ADDR"`
		Password string `yaml:"password" env:"DATA_REDIS_PASSWORD"`
	}
	Log struct {
		LogPath      string `yaml:"log_path"`
		LogLevel     string `yaml:"log_level"`
		LogEncodeMod string `yaml:"log_encode_mod"`
		IsConsole    bool   `yaml:"is_console"`
	}
	Email struct {
		Sender string `yaml:"sender"` //发送人邮箱（邮箱以自己的为准）
		Name   string `yaml:"name"`
		Pass   string `yaml:"pass"` //发送人邮箱的密码，现在可能会需要邮箱 开启授权密码后在pass填写授权码
		Host   string `yaml:"host"` //邮箱服务器（此时用的是qq邮箱）
		Port   int    `yaml:"port"` //邮箱服务器端口
	}
	Auth struct {
		Secret string `yaml:"secret"`
	}
	AliOss struct {
		AccessKeyId     string `yaml:"access_key_id" env:"THIRD_PARTY_ALI_OSS_ACCESS_KEY_ID"`
		AccessKeySecret string `yaml:"access_key_secret" env:"THIRD_PARTY_ALI_OSS_ACCESS_KEY_SECRET"`
		Endpoint        string `yaml:"endpoint" env:"THIRD_PARTY_ALI_OSS_ENDPOINT"`
	}
)

var Bootstrap = &Config{}

func init() {
	err := cleanenv.ReadConfig("D:\\Goland\\go_xm\\lalala_im\\configs\\receive_server_conf.yml", Bootstrap)
	if err != nil {
		panic(errors.Wrap(err, "初始化配置文件错误"))
	}
	if Bootstrap.App.RunMode != "debug" {
		//初始化日志
		log.InitLog(Bootstrap.Log.LogPath, Bootstrap.Log.LogLevel, Bootstrap.Log.LogEncodeMod, Bootstrap.Log.IsConsole)
	} else {
		//初始化日志
		log.InitLog("", "", "", true)
	}
	db.InitDB(Bootstrap.Data.Mongo.Url, Bootstrap.Data.Mongo.Database, Bootstrap.Data.Redis.ADDR, Bootstrap.Data.Redis.Password,
		Bootstrap.AliOss.Endpoint, Bootstrap.AliOss.AccessKeyId, Bootstrap.AliOss.AccessKeySecret)
	la_email.InitEmail(Bootstrap.Email.Sender, Bootstrap.Email.Name, Bootstrap.Email.Pass, Bootstrap.Email.Host, Bootstrap.Email.Port)
	la_auth.InitSecret(Bootstrap.App.Auth.Secret)
}
