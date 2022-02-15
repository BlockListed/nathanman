package config

import (
	"log"
	"os"
)

type DbConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SslMode  string
}

type DiscordConfig struct {
	Token         string
	ListenChannel string
	SendChannel   string
	UserId        string
	GuildId       string
}

type Config struct {
	Discord  DiscordConfig
	Database DbConfig
	Timezone string
}

var Logger = log.New(os.Stderr, "[NathanMan] ", log.Ltime|log.Ldate|log.Lmicroseconds|log.Lshortfile)

var Configuration = InitiateConfig()

func InitiateConfig() (r Config) {
	r.Database = InitiateDbConfig()
	r.Timezone = loadEnvvar("TZ", "Etc/UTC")
	r.Discord = initiateDiscordConfig()
	return
}

func initiateDiscordConfig() (r DiscordConfig) {
	r.Token = panicEnvvar("DISCORD_TOKEN")
	r.ListenChannel = panicEnvvar("DISCORD_LISTEN_CHANNEL")
	r.SendChannel = panicEnvvar("DISCORD_SEND_CHANNEL")
	r.UserId = panicEnvvar("DISCORD_USER_ID")
	r.GuildId = panicEnvvar("DISCORD_GUILD_ID")
	return
}

func InitiateDbConfig() (r DbConfig) {
	r.Host = loadEnvvar("DB_HOST", "localhost")
	r.Port = loadEnvvar("DB_PORT", "5432")
	r.Name = loadEnvvar("DB_NAME", "nathanman")
	r.User = loadEnvvar("DB_USER", "nathanman")
	r.Password = loadEnvvar("DB_PASSWORD", "nathanman")
	r.SslMode = loadEnvvar("DB_SSL", "disable")
	return
}

func loadEnvvar(k string, default_return string) string {
	v, b := os.LookupEnv(k)
	if b {
		return v
	} else {
		return default_return
	}
}

func panicEnvvar(k string) string {
	v, b := os.LookupEnv(k)
	if !b || v == "" {
		log.New(os.Stderr, "[NathanMan]", log.Lmsgprefix|log.Ltime|log.Ldate|log.Lmicroseconds|log.Lshortfile).Panicf("Couldn't get required env var: %s!", k)
		return ""
	} else {
		return v
	}
}
