package spine

import (
	"errors"
	"log"
	"math/rand"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
)

type SpineService struct {
	Assets *AssetService
}

func NewSpineService(assets *AssetService) *SpineService {
	return &SpineService{
		Assets: assets,
	}
}

func (s *SpineService) ValidateOperatorRequest(info *OperatorInfo) error {
	assetMap := s.Assets.getAssetMapFromFaction(info.Faction)

	log.Println("Request setOperator", info.OperatorId, info.Faction,
		info.Skin, info.ChibiStance, info.Facing, info.CurrentAction)
	err := assetMap.Contains(info.OperatorId, info.Skin, info.ChibiStance,
		info.Facing, info.Action.GetAnimations(info.CurrentAction))
	if err != nil {
		log.Println("Validate setOperator request failed", err)
		return err
	}
	return nil
}

func (s *SpineService) GetSpineData(opeatorId string, faction FactionEnum, skin string, isBase bool, isFront bool) *SpineData {
	assetMap := s.Assets.getAssetMapFromFaction(faction)
	return assetMap.Get(opeatorId, skin, isBase, isFront)
}

func (s *SpineService) GetOperatorIds(faction FactionEnum) ([]string, error) {
	assetMap := s.Assets.getAssetMapFromFaction(faction)
	operatorIds := make([]string, 0)
	for operatorId := range assetMap.Data {
		operatorIds = append(operatorIds, operatorId)
	}
	return operatorIds, nil
}

func (s *SpineService) GetOperator(
	OperatorId string,
	Faction FactionEnum,
) (*GetOperatorResponse, error) {
	assetMap := s.Assets.getAssetMapFromFaction(Faction)

	operatorData, ok := assetMap.Data[OperatorId]
	if !ok {
		return nil, errors.New("No operator with id " + OperatorId + " is loaded")
	}

	skinDataMap := make(map[string]SkinData)
	for skinName, skin := range operatorData.Skins {

		baseMapping := make(map[ChibiFacingEnum]AnimationsList, 0)
		battleMapping := make(map[ChibiFacingEnum]AnimationsList, 0)
		for facing, spineData := range skin.Base {
			baseMapping[facing] = spineData.Animations
		}
		for facing, spineData := range skin.Battle {
			battleMapping[facing] = spineData.Animations
		}

		skinDataMap[skinName] = SkinData{
			Stances: map[ChibiStanceEnum]FacingData{
				CHIBI_STANCE_ENUM_BASE: {
					Facings: baseMapping,
				},
				CHIBI_STANCE_ENUM_BATTLE: {
					Facings: battleMapping,
				},
			},
		}
	}

	canonicalName := s.Assets.getCommonNamesFromFaction(Faction).GetCanonicalName(OperatorId)

	return &GetOperatorResponse{
		OperatorId:   OperatorId,
		OperatorName: canonicalName,
		Skins:        skinDataMap,
	}, nil
}

func (s *SpineService) GetOperatorIdFromName(name string, faction FactionEnum) (string, []string) {
	commonNames := s.Assets.getCommonNamesFromFaction(faction)

	if operatorId, ok := commonNames.IsMatch(name); ok {
		return operatorId, nil
	}

	matches := commonNames.FindMatchs(name, 5)
	humanMatches := make([]string, 0)
	for _, match := range matches {
		humanMatches = append(humanMatches, commonNames.operatorIdToNames[match][0])
	}
	return "", humanMatches
}

func (s *SpineService) GetRandomOperator() (*OperatorInfo, error) {
	operatorIds, err := s.GetOperatorIds(FACTION_ENUM_OPERATOR)
	if err != nil {
		return nil, err
	}

	index := rand.Intn(len(operatorIds))
	operatorId := operatorIds[index]

	faction := FACTION_ENUM_OPERATOR
	operatorData, err := s.GetOperator(operatorId, faction)
	if err != nil {
		return nil, err
	}
	chibiStance := CHIBI_STANCE_ENUM_BASE
	skinName := DEFAULT_SKIN_NAME
	stanceMap, ok := operatorData.Skins[skinName].Stances[chibiStance]
	if !ok {
		chibiStance = CHIBI_STANCE_ENUM_BATTLE
	}
	if len(stanceMap.Facings) == 0 {
		chibiStance = CHIBI_STANCE_ENUM_BATTLE
	}
	facing := CHIBI_FACING_ENUM_FRONT
	availableAnimations := operatorData.Skins[skinName].Stances[chibiStance].Facings[facing]
	availableAnimations = FilterAnimations(availableAnimations)
	availableSkins := operatorData.GetSkinNames()

	commonNames := s.Assets.getCommonNamesFromFaction(faction)
	operatorDisplayName := commonNames.GetCanonicalName(operatorId)

	opInfo := NewOperatorInfo(
		operatorDisplayName,
		faction,
		operatorId,
		skinName,
		chibiStance,
		facing,
		availableSkins,
		availableAnimations,
		1.0,
		misc.EmptyOption[misc.Vector2](),
		ACTION_PLAY_ANIMATION,
		NewActionPlayAnimation(
			[]string{GetDefaultAnimForChibiStance(chibiStance)},
		),
	)
	return &opInfo, nil
}

func (s *SpineService) OperatorFromDefault(
	opName string,
	details misc.InitialOperatorDetails,
) *OperatorInfo {
	if len(opName) == 0 {
		opName = "Amiya"
	}

	faction := FACTION_ENUM_OPERATOR
	opId, matches := s.GetOperatorIdFromName(opName, FACTION_ENUM_OPERATOR)
	if matches != nil {
		faction = FACTION_ENUM_ENEMY
		opId, matches = s.GetOperatorIdFromName(opName, FACTION_ENUM_ENEMY)
	}
	if matches != nil {
		log.Panic("Failed to get operator id", matches)
	}
	stance, err2 := ChibiStanceEnum_Parse(details.Stance)
	if err2 != nil {
		log.Panic("Failed to parse stance", err2)
	}

	opResp, err := s.GetOperator(opId, faction)
	if err != nil {
		log.Panic("Failed to fetch operator info")
	}
	availableAnims := opResp.Skins[details.Skin].Stances[stance].Facings[CHIBI_FACING_ENUM_FRONT]
	availableAnims = FilterAnimations(availableAnims)
	availableSkins := opResp.GetSkinNames()

	opInfo := NewOperatorInfo(
		opResp.OperatorName,
		faction,
		opId,
		details.Skin,
		stance,
		CHIBI_FACING_ENUM_FRONT,
		availableSkins,
		availableAnims,
		1.0,
		misc.NewOption(misc.Vector2{X: details.PositionX, Y: 0.0}),
		ACTION_PLAY_ANIMATION,
		NewActionPlayAnimation(details.Animations),
	)
	return &opInfo
}
