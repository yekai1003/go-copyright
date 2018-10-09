package configs

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

var (
	Version   = "1.0.0"
	Commit    = ""
	BuildTime = "2018-10-09"
)

type ServerConfig struct {
	Common *CommonConfig
}

type CommonConfig struct {
	Port      string
	LogFormat string
}

type DbConfig struct {
	Driver string
	Url    string
}

func usage() {
	fmt.Printf("Usage: %s -c config_file [-v] [-h]\n", os.Args[0])
}

func GetConfig() (config *ServerConfig) {
	var configFile = flag.String("c", "", "Config file")

	var ver = flag.Bool("v", false, "version")
	var help = flag.Bool("h", false, "Help")

	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		return nil
	}

	if *ver {
		fmt.Println("Version: ", Version)
		fmt.Println("Commit: ", Commit)
		fmt.Println("BuildTime: ", BuildTime)
		return nil
	}
	// get server config
	if *configFile != "" {
		config = &ServerConfig{}
		if _, err := toml.DecodeFile(*configFile, &config); err != nil {
			panic(err)
		}
	} else {
		usage()
		return nil
	}

	return config
}
