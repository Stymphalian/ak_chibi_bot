package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

func run(assetDir string, outputDir string) error {
	assetMap := operator.NewSpineAssetMap()
	customMap := operator.NewSpineAssetMap()
	enemyAssetMap := operator.NewSpineAssetMap()

	if err := assetMap.Load(assetDir, "characters"); err != nil {
		return err
	}
	if err := customMap.Load(assetDir, "custom"); err != nil {
		return err
	}
	if err := enemyAssetMap.Load(assetDir, "enemies"); err != nil {
		return err
	}

	assetMapJsonBytes, err := json.MarshalIndent(assetMap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "characters_index.json"), assetMapJsonBytes, 0666); err != nil {
		return err
	}

	customMapJsonBytes, err := json.MarshalIndent(customMap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "custom_index.json"), customMapJsonBytes, 0666); err != nil {
		return err
	}

	enemyMapJsonBytes, err := json.MarshalIndent(enemyAssetMap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "enemy_index.json"), enemyMapJsonBytes, 0666); err != nil {
		return err
	}
	return nil
}

// cd server/tools/index
// go run create_asset_index.go -assetDir ../../../static/assets -outputDir .
// go run create_asset_index.go -assetDir C:/dev/lab/ak_etl/output/organized -outputDir C:/dev/lab/ak_etl/output/organized
func main() {
	assetDirPtr := flag.String("assetDir", "", "path to the assets")
	outputDirPtr := flag.String("outputDir", "", "path to the output directory")
	flag.Parse()
	if *assetDirPtr == "" {
		log.Fatal("must specify -assetDir")
	}
	if *outputDirPtr == "" {
		log.Fatal("must specify -outputDir")
	}

	log.Println("-assetDir: ", *assetDirPtr)
	log.Println("-outputDir: ", *outputDirPtr)
	log.Println(run(*assetDirPtr, *outputDirPtr))
}
