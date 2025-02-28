package common

import (
	"flag"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

// Put all env variables here

var (
	secretsJsonFile string
	discordJsonFile string
)

var (
	MainBotToken     string
	ServerBotToken   string
	ClientSecret     string
	MainServer       string
	TestServer       string
	StaffServer      string
	SiteLogs         string
	CertifiedBotRole string
	CertifiedDevRole string
	BotDevRole       string
	CliCmd           string
	RootPath         string
	PythonPath       string
)

func init() {
	flag.StringVar(&secretsJsonFile, "secret", "/home/meow/FatesList/config/data/secrets.json", "Secrets json file")
	flag.StringVar(&discordJsonFile, "discord", "/home/meow/FatesList/config/data/discord.json", "Discord json file")
	flag.StringVar(&staffRoleFilePath, "staff-roles", "/home/meow/FatesList/config/data/staff_roles.json", "Staff roles json")
	flag.StringVar(&CliCmd, "cmd", "", "The command to run:\n\tdragon.server: the dragon ipc and ws server\n\tsite.XXX: run a site command (run=run site, compilestatic=compile static files).\n\tSet PYLOG_LEVEL to set loguru log level to debug")
	flag.StringVar(&RootPath, "root", "/home/meow/FatesList", "Fates List source directory")
	flag.StringVar(&PythonPath, "python-path", "/home/meow/flvenv-py11/bin/python", "Path to python interpreter")
	flag.Parse()

	if CliCmd == "" {
		flag.Usage()
		os.Exit(3)
	}

	err := os.Chdir(RootPath)

	if err != nil {
		panic(err)
	}

	var secretsJson, ferr = ioutil.ReadFile(secretsJsonFile)
	var discordJson, ferr2 = ioutil.ReadFile(discordJsonFile)
	if ferr != nil {
		panic(ferr.Error())
	} else if ferr2 != nil {
		panic(ferr2.Error())
	}

	MainBotToken = fastjson.GetString(secretsJson, "token_main")
	ClientSecret = fastjson.GetString(secretsJson, "client_secret")
	ServerBotToken = fastjson.GetString(secretsJson, "token_server")

	var p fastjson.Parser

	v, err := p.Parse(string(discordJson))

	if err != nil {
		panic(err)
	}

	var servers = v.GetObject("servers")

	MainServer = string(servers.Get("main").GetStringBytes())
	TestServer = string(servers.Get("testing").GetStringBytes())
	StaffServer = string(servers.Get("staff").GetStringBytes())

	var channels = v.GetObject("channels")

	SiteLogs = string(channels.Get("bot_logs").GetStringBytes())

	var roles = v.GetObject("roles")

	CertifiedBotRole = string(roles.Get("certified_bots_role").GetStringBytes())
	CertifiedDevRole = string(roles.Get("certified_dev_role").GetStringBytes())
	BotDevRole = string(roles.Get("bot_dev_role").GetStringBytes())

	permInit()

	log.Info("Environment setup successfully!")
}
