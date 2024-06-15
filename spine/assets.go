package spine

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

const (
	DEFAULT_ANIM_BASE      = "Relax"
	DEFAULT_ANIM_BATTLE    = "Idle"
	DEFAULT_SKIN_NAME      = "default"
	DEFAULT_MOVE_ANIM_NAME = "Move"
)

type ChibiFacingEnum string

const (
	CHIBI_FACING_ENUM_FRONT ChibiFacingEnum = "Front"
	CHIBI_FACING_ENUM_BACK  ChibiFacingEnum = "Back"
)

func ChibiFacingEnum_Parse(str string) (ChibiFacingEnum, error) {
	switch strings.ToLower(str) {
	case "front":
		return CHIBI_FACING_ENUM_FRONT, nil
	case "back":
		return CHIBI_FACING_ENUM_BACK, nil
	default:
		return "", fmt.Errorf("invalid chibi facing (%s)", str)
	}
}

type ChibiTypeEnum string

const (
	CHIBI_TYPE_ENUM_BATTLE ChibiTypeEnum = "battle"
	CHIBI_TYPE_ENUM_BASE   ChibiTypeEnum = "base"
)

func ChibiTypeEnum_Parse(str string) (ChibiTypeEnum, error) {
	switch strings.ToLower(str) {
	case "battle":
		return CHIBI_TYPE_ENUM_BATTLE, nil
	case "base":
		return CHIBI_TYPE_ENUM_BASE, nil
	default:
		return "", fmt.Errorf("invalid chibi type (%s)", str)
	}
}

type FactionEnum string

const (
	FACTION_ENUM_OPERATOR FactionEnum = "operator"
	FACTION_ENUM_ENEMY    FactionEnum = "enemy"
)

func FactionEnum_Parse(str string) (FactionEnum, error) {
	switch strings.ToLower(str) {
	case "operator":
		return FACTION_ENUM_OPERATOR, nil
	case "enemy":
		return FACTION_ENUM_ENEMY, nil
	default:
		return "", fmt.Errorf("invalid faction type (%s)", str)
	}
}

// TODO:
func GetDefaultAnimForChibiType(chibiType ChibiTypeEnum) string {
	if chibiType == CHIBI_TYPE_ENUM_BASE {
		return DEFAULT_ANIM_BASE
	} else {
		return DEFAULT_ANIM_BATTLE
	}
}

type SpineAssetMap struct {
	// map[operatorId]*ChibiAssetPathEntry
	// map["char_002_amiya"]*ChibiAssetPathEntry
	Data map[string]*ChibiAssetPathEntry `json:"data"`
}
type ChibiAssetPathEntry struct {
	// map[skinName]*SpineSkinData
	Skins map[string]*SpineSkinData `json:"skins"`
}
type SpineSkinData struct {
	//map[Front,Back]*SpineData
	Base   map[ChibiFacingEnum]*SpineData `json:"base"`
	Battle map[ChibiFacingEnum]*SpineData `json:"battle"`
}
type SpineData struct {
	AtlasFilepath string `json:"AtlasFilepath"`
	SkelFilepath  string `json:"AtlasFullFilepath"`
	PngFilepath   string `json:"png_filepath"`

	AtlasFullFilepath string `json:"atlas_full_filepath"`
	SkelFullFilepath  string `json:"skel_full_filepath"`
	PngFullFilepath   string `json:"png_full_filepath"`

	Animations []string `json:"animations"`
}

func NewSpineAssetMap() *SpineAssetMap {
	return &SpineAssetMap{
		Data: make(map[string]*ChibiAssetPathEntry),
	}
}

func readJsonSkelAnimations(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make(map[string][]string)

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {

		return nil, err
	}

	return data["animations"], nil
}

func (s *SpineAssetMap) Load(assetDir string, assetSubdir string) error {
	log.Println("Loading Asset maps")
	charsDir := filepath.Join(assetDir, assetSubdir)

	err := filepath.Walk(charsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		newPath := strings.TrimPrefix(path, assetDir+string(os.PathSeparator))
		pathList := strings.Split(newPath, string(os.PathSeparator))

		operatorName := pathList[1] // char_002_amiya
		skinName := pathList[2]     // epoque#4,default
		battleOrBase := pathList[3] // battle or base
		frontOrBack := pathList[4]  // Front or Back
		spineData := s.Get(operatorName, skinName, battleOrBase == "base", frontOrBack == "Front")

		switch filepath.Ext(info.Name()) {
		case ".atlas":
			spineData.AtlasFilepath = newPath
			spineData.AtlasFullFilepath = path
		case ".png":
			spineData.PngFilepath = newPath
			spineData.PngFullFilepath = path
		case ".skel":
			spineData.SkelFilepath = newPath
			spineData.SkelFullFilepath = path
		case ".json":
			animations, err := readJsonSkelAnimations(path)
			if err != nil {
				log.Fatal("Failed to extract animations from skeleton json: ", err)
			}
			spineData.Animations = animations
		default:
			log.Fatal("Unknown extension: ", info.Name())
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Fix up the skin names to make them easier to work with
	for _, entry := range s.Data {

		for skinName := range entry.Skins {
			newSkinName := skinName
			if strings.HasPrefix(skinName, "ambienceSynesthesia") {
				newSkinName = strings.ReplaceAll(skinName, "ambienceSynesthesia", "ambience")
			}

			// Check to see if we can remove the #\d number from the
			// the end of the skin name. If there are multiple skins from the
			// same family (ie. epoque#4 and epoque#5) then we just remove the
			// the '#' symbol but leave the number in place
			hasDuplicate := false
			numRemoved := regexp.MustCompile(`#\d+$`).ReplaceAllString(newSkinName, "")
			for otherSkinName := range entry.Skins {
				if otherSkinName == skinName {
					continue
				}

				if strings.HasPrefix(otherSkinName, numRemoved) {
					hasDuplicate = true
					break
				}
			}
			if !hasDuplicate {
				newSkinName = numRemoved
			} else {
				newSkinName = strings.ReplaceAll(newSkinName, "#", "")
			}

			if newSkinName != skinName {
				// fmt.Println("Skin: ", skinName, " -> ", newSkinName, " ", operatorName)
				entry.Skins[newSkinName] = entry.Skins[skinName]
				delete(entry.Skins, skinName)
			}
		}
	}

	fmt.Printf("Loaded %d assets from %s\n", len(s.Data), assetDir)
	return nil

	// for opName, entry := range assetMap.data {
	// 	log.Println(opName)
	// 	for skinName, skin := range entry.skins {
	// 		log.Printf("  %v", skinName)
	// 		log.Printf("    base")
	// 		for facingDir, data := range skin.base {
	// 			log.Printf("      %v", facingDir)
	// 			log.Printf("        %v", data.AtlasFilepath)
	// 			log.Printf("        %v", data.png_filepath)AtlasFullFilepath
	// 			log.Printf("        %v", data.skel_filepath)
	// 		}
	// 		log.Printf("    battle")
	// 		for facingDir, data := range skin.battle {
	// 			log.Printf("      %v", facingDir)
	// 			log.Printf("        %v", data.AtlasFilepath)
	// 			log.Printf("        %v", data.png_filepath)AtlasFullFilepath
	// 			log.Printf("        %v", data.skel_filepath)
	// 		}
	// 	}
	// }
}

func (s *SpineAssetMap) Get(
	operatorId string,
	skin string,
	isBase bool,
	isFront bool) *SpineData {
	frontOrBack := CHIBI_FACING_ENUM_BACK
	if isFront {
		frontOrBack = CHIBI_FACING_ENUM_FRONT
	}

	if _, ok := s.Data[operatorId]; !ok {
		s.Data[operatorId] = &ChibiAssetPathEntry{
			Skins: make(map[string]*SpineSkinData),
		}
	}
	if _, ok := s.Data[operatorId].Skins[skin]; !ok {
		s.Data[operatorId].Skins[skin] = &SpineSkinData{
			Base:   make(map[ChibiFacingEnum]*SpineData),
			Battle: make(map[ChibiFacingEnum]*SpineData),
		}
	}
	if isBase {
		if _, ok := s.Data[operatorId].Skins[skin].Base[frontOrBack]; !ok {
			s.Data[operatorId].Skins[skin].Base[frontOrBack] = &SpineData{}
		}
		return s.Data[operatorId].Skins[skin].Base[frontOrBack]
	} else {
		if _, ok := s.Data[operatorId].Skins[skin].Battle[frontOrBack]; !ok {
			s.Data[operatorId].Skins[skin].Battle[frontOrBack] = &SpineData{}
		}
		return s.Data[operatorId].Skins[skin].Battle[frontOrBack]
	}
}

func (s *SpineAssetMap) Contains(operatorId string, skin string, chibiType ChibiTypeEnum, facing ChibiFacingEnum, animations []string) error {
	if _, ok := s.Data[operatorId]; !ok {
		return fmt.Errorf("invalid operator name (%s)", operatorId)
	}
	if _, ok := s.Data[operatorId].Skins[skin]; !ok {
		return fmt.Errorf("invalid skin name (%s)", skin)
	}
	var spineData *SpineData
	if chibiType == CHIBI_TYPE_ENUM_BASE {
		if len(s.Data[operatorId].Skins[skin].Base) == 0 {
			return fmt.Errorf("skin does not have a 'Base' skin")
		}
		if _, ok := s.Data[operatorId].Skins[skin].Base[facing]; !ok {
			return fmt.Errorf("base skin does not have facing drection (%s)", facing)
		}
		spineData = s.Data[operatorId].Skins[skin].Base[facing]
	} else {
		if len(s.Data[operatorId].Skins[skin].Battle) == 0 {
			return fmt.Errorf("skin does not have a 'Battle' skin")
		}
		if _, ok := s.Data[operatorId].Skins[skin].Battle[facing]; !ok {
			return fmt.Errorf("battle skin does not have facing drection (%s)", facing)
		}
		spineData = s.Data[operatorId].Skins[skin].Battle[facing]
	}

	for _, animation := range animations {
		if !slices.Contains(spineData.Animations, animation) {
			return fmt.Errorf("skin does not have animation (%s)", animation)
		}
	}

	return nil
}

type CommonNames struct {
	operatorIdToNames map[string]([]string)
	namesToOperatorId map[string]([]string)
	allNames          []string
}

func NewCommonNames() *CommonNames {
	return &CommonNames{
		operatorIdToNames: make(map[string]([]string)),
		namesToOperatorId: make(map[string]([]string)),
		allNames:          make([]string, 0),
	}
}

func (s *CommonNames) Load(assetFilePath string) error {
	savedNames := make(map[string]([]string))
	data, err := os.ReadFile(assetFilePath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(data, &savedNames)
	if err != nil {
		log.Fatal(err)
	}

	nameToOperatorId := make(map[string][]string)
	allNames := make([]string, len(savedNames))

	dupes := make([]string, 0)
	for operatorId, names := range savedNames {
		for _, name := range names {
			if _, ok := nameToOperatorId[name]; !ok {
				nameToOperatorId[name] = make([]string, 0)
			}
			if len(nameToOperatorId[name]) > 0 {
				dupes = append(dupes, name)
			}
			nameToOperatorId[name] = append(nameToOperatorId[name], operatorId)
			allNames = append(allNames, name)
		}
	}

	if len(dupes) > 0 {
		//A7, EB, SA1, SA2
		for _, name := range dupes {
			log.Printf("Duplicate names: %s\n", name)
		}
	}

	s.operatorIdToNames = savedNames
	s.namesToOperatorId = nameToOperatorId
	s.allNames = allNames
	log.Printf("Found %d common names\n", len(s.operatorIdToNames))
	return nil
	// go func() {
	// 	scanner := bufio.NewScanner(os.Stdin)
	// 	for scanner.Scan() {
	// 		text := scanner.Text()

	// 		matches := s.commonNames.FindMatchs(text, 3)
	// 		log.Println("Did you mean...")
	// 		for _, m := range matches {
	// 			log.Println(s.commonNames.operatorIdToNames[m][0])
	// 		}

	// 	}
	// 	if err := scanner.Err(); err != nil {
	// 		log.Println("read stdin:", err)
	// 		return
	// 	}
	// }()
}

func (s *CommonNames) GetCanonicalName(operatorId string) string {
	if names, ok := s.operatorIdToNames[operatorId]; ok {
		return names[0]
	}
	log.Printf("invalid operatorId: %s", operatorId)
	return ""
}

func (s *CommonNames) IsMatch(name string) (string, bool) {
	if operatorIds, ok := s.namesToOperatorId[name]; ok {
		if len(operatorIds) != 1 {
			return "", false
		}
		return operatorIds[0], true
	}
	// TODO: Improve this O(n) algorithm
	for operatorName, operatorIds := range s.namesToOperatorId {
		if strings.EqualFold(operatorName, name) {
			if len(operatorIds) != 1 {
				return "", false
			}
			return operatorIds[0], true
		}
	}
	return "", false
}

func (s *CommonNames) FindMatchs(userInput string, numSuggestions int) (output []string) {
	rankedMatches := fuzzy.RankFindNormalizedFold(userInput, s.allNames)
	sort.Sort(rankedMatches)

	set := make(map[string]bool)
	out := false
	for _, match := range rankedMatches {
		log.Printf("Found multiple matches: %v\n", match)
		operatorIds := s.namesToOperatorId[match.Target]
		for _, operatorId := range operatorIds {

			if _, ok := set[operatorId]; !ok {
				set[operatorId] = true
				output = append(output, operatorId)
			}
			if len(output) >= numSuggestions {
				out = true
				break
			}
		}
		if out {
			break
		}
	}
	return
}
