package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/Stymphalian/ak_chibi_bot/spine"
)

func main() {
	assetDir := flag.String("assetdir", "/ak_chibi_assets/assets", "Asset directory")
	flag.Parse()

	assetMap := spine.NewSpineAssetMap()
	err := assetMap.Load(*assetDir, "characters")
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(assetMap, "", "  ")
	newAssetMap := spine.NewSpineAssetMap()
	json.Unmarshal(out, &newAssetMap)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(newAssetMap)
}
