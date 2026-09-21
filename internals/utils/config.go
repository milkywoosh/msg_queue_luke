package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	AllowedOrigins       []string      `mapstructure:"ALLOWED_ORIGINS"`
	DBSource             string        `mapstructure:"PG_CONNSTRING"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	RedisPassword        string        `mapstructure:"REDIS_PASSWORD"`
	RabbitMQDial         string        `mapstructure:"RABBITMQ_DIAL"`
	RabbitMQUser         string        `mapstructure:"RABBITMQ_USER"`
	RabbitMQPassword     string        `mapstructure:"RABBITMQ_PASSWORD"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION_INMINUTE"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	Port                 string        `mapstructure:"PORT"`
	Addr                 string        `mapstructure:"ADDR"`
	AccessKeyS3          string        `mapstructure:"ACCESS_KEY_S3"`
	SecretKeyS3          string        `mapstructure:"SECRET_KEY_S3"`
	AddressS3            string        `mapstructure:"ENDPOINT_S3"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	fields := []string{
		"ENVIRONMENT", "ALLOWED_ORIGINS", "PG_CONNSTRING", "MIGRATION_URL",
		"REDIS_ADDRESS", "REDIS_PASSWORD", "HTTP_SERVER_ADDRESS", "GRPC_SERVER_ADDRESS",
		"TOKEN_SYMMETRIC_KEY", "ACCESS_TOKEN_DURATION_INMINUTE", "REFRESH_TOKEN_DURATION",
		"EMAIL_SENDER_NAME", "EMAIL_SENDER_ADDRESS", "EMAIL_SENDER_PASSWORD",
		"PORT", "ADDR", "RABBITMQ_USER", "RABBITMQ_PASSWORD", "RABBITMQ_DIAL",
		"ACCESS_KEY_S3", "SECRET_KEY_S3", "ENDPOINT_S3",
	}
	for _, f := range fields {
		viper.BindEnv(f)
	}

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
		// not found is fine — fall through to AutomaticEnv()
	}

	err = viper.Unmarshal(&config)
	return
}
