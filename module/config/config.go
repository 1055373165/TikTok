package config

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var config *viper.Viper

type MySQL struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

type S3 struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`

	// 如果是使用 minio，并且没有使用 https，需要设置为 true
	UseSsl *bool `yaml:"useSsl"`
	// 如果是使用 minio，需要设置为 true
	HostnameImmutable *bool `yaml:"hostnameImmutable"`
}

var (
	mysql    MySQL
	redis    Redis
	S3Config S3
)

func Init() {
	config = viper.New()

	config.SetConfigName("tiktok")
	config.AddConfigPath("./conf/")
	config.AddConfigPath("./")
	config.AddConfigPath("$HOME/.tiktok/")
	config.AddConfigPath("/etc/tiktok/")
	config.SetConfigType("yml")

	config.AutomaticEnv()
	config.SetEnvPrefix("TK")
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)

	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件发生变更之后会调用的回调函数
		fmt.Println("Config file changed:", e.Name)
	})

	if err := config.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// 配置文件未找到错误
			fmt.Println("config file not found use default config")
			config.SetDefault("server", map[string]interface{}{
				"port": 8080,
				"mode": "prod",
			})

			config.SetDefault("log", map[string]interface{}{
				"level": "debug",
				"mode":  []string{"console", "file"},
				"path":  "./log",
			})

			config.SetDefault("mysql", map[string]interface{}{
				"host":     "127.0.0.1",
				"port":     3306,
				"username": "root",
				"password": "123456",
				"database": "tiktok",
			})

			config.SetDefault("redis", map[string]interface{}{
				"host":     "127.0.0.1",
				"port":     6379,
				"password": "123456",
				"database": 0,
			})
		}
	}
	err := config.UnmarshalKey("mysql", &mysql)
	if err != nil {
		log.Fatalf("unable to decode into mysql struct, %v", err)
	}
	err = config.UnmarshalKey("redis", &redis)
	if err != nil {
		log.Fatalf("unable to decode into redis struct, %v", err)
	}
	err = config.UnmarshalKey("s3", &S3Config)
	if err != nil {
		log.Fatalf("unable to decode into s3 struct, %v", err)
	}

	// use env var to set S3 config when some field is nil
	if S3Config.Endpoint == "" {
		S3Config.Endpoint = config.GetString("s3.endpoint")
	}
	if S3Config.AccessKey == "" {
		S3Config.AccessKey = config.GetString("s3.accesskey")
	}
	if S3Config.SecretKey == "" {
		S3Config.SecretKey = config.GetString("s3.secretkey")
	}
	if S3Config.Region == "" {
		S3Config.Region = config.GetString("s3.region")
	}
	if S3Config.Bucket == "" {
		S3Config.Bucket = config.GetString("s3.bucket")
	}
}

func Get(key string) interface{} {
	return config.Get(key)
}

func GetString(key string) string {
	return config.GetString(key)
}

func MySQLDSN() string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v"+
		"?charset=utf8mb4&parseTime=True&loc=Local",
		mysql.Username, mysql.Password, mysql.Host,
		mysql.Port, mysql.Database)
}
