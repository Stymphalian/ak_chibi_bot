package misc

import (
	"errors"
	"flag"
	"log"
)

type CommandLineArgs struct {
	ImageAssetDir  string
	StaticAssetDir string
	Address        string
	BotConfigPath  string
	BotConfig      BotConfig
}

func ProvideCommandLineArgs() (*CommandLineArgs, error) {
	log.Println("Providing CommandLineArgs")
	imageAssetDir := flag.String("image_assetdir", "../static/assets", "Image Asset Directory")
	staticAssetsDir := flag.String("static_dir", "../static", "Static assets folder")
	address := flag.String("address", ":8080", "Server address")
	botConfigPath := flag.String("bot_config", "bot_config.json", "Config filepath containing channel names and tokens")
	flag.Parse()
	log.Println("-image_assetdir: ", *imageAssetDir)
	log.Println("-static_dir: ", *staticAssetsDir)
	log.Println("-address: ", *address)
	log.Println("-bot_config:", *botConfigPath)

	if *botConfigPath == "" {
		return nil, errors.New("must specify -bot_config")
	}

	botConfig, err := LoadBotConfig(*botConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	return &CommandLineArgs{
		ImageAssetDir:  *imageAssetDir,
		StaticAssetDir: *staticAssetsDir,
		Address:        *address,
		BotConfigPath:  *botConfigPath,
		BotConfig:      *botConfig,
	}, nil
}

type ImageAssetDirString string
type StaticAssetDirString string
type AddressString string
type BotConfigPath string

func ProvideImageAssetDirString(args *CommandLineArgs) ImageAssetDirString {
	return ImageAssetDirString(args.ImageAssetDir)
}

func ProvideStaticAssetDirString(args *CommandLineArgs) StaticAssetDirString {
	return StaticAssetDirString(args.StaticAssetDir)
}

func ProvideAddressString(args *CommandLineArgs) AddressString {
	return AddressString(args.Address)
}

func ProvideBotConfigPath(args *CommandLineArgs) BotConfigPath {
	return BotConfigPath(args.BotConfigPath)
}

func ProvideBotConfig(args *CommandLineArgs) *BotConfig {
	return &args.BotConfig
}
