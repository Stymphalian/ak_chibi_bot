package spine

import (
	"errors"
	"log"
	"math/rand"
	"slices"

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

func (c *SpineService) ValidateUpdateSetDefaultOtherwise(update *OperatorInfo) error {
	if len(update.Faction) == 0 {
		update.Faction = FACTION_ENUM_OPERATOR
	}

	currentOp, err := c.GetOperator(update.OperatorId, update.Faction)
	if err != nil {
		return errors.New("something went wrong please try again")
	}
	update.OperatorDisplayName = currentOp.OperatorName

	if _, ok := currentOp.Skins[update.Skin]; !ok {
		update.Skin = DEFAULT_SKIN_NAME
	}
	facings := currentOp.Skins[update.Skin].Stances[update.ChibiStance]
	if len(facings.Facings) == 0 {
		switch update.ChibiStance {
		case CHIBI_STANCE_ENUM_BASE:
			update.ChibiStance = CHIBI_STANCE_ENUM_BATTLE
		case CHIBI_STANCE_ENUM_BATTLE:
			update.ChibiStance = CHIBI_STANCE_ENUM_BASE
		default:
			update.ChibiStance = CHIBI_STANCE_ENUM_BASE
		}
	}
	if _, ok := currentOp.Skins[update.Skin].Stances[update.ChibiStance].Facings[update.Facing]; !ok {
		update.Facing = CHIBI_FACING_ENUM_FRONT
	}

	update.Skins = currentOp.GetSkinNames()

	// Validate animations
	update.AvailableAnimations = currentOp.Skins[update.Skin].Stances[update.ChibiStance].Facings[update.Facing]
	update.AvailableAnimations = FilterAnimations(update.AvailableAnimations)

	// Validate animationSpeed
	if update.AnimationSpeed == 0 {
		update.AnimationSpeed = DEFAULT_ANIMATION_SPEED
	}
	update.AnimationSpeed = misc.ClampF64(
		update.AnimationSpeed,
		MIN_ANIMATION_SPEED,
		MAX_ANIMATION_SPEED,
	)

	// Sprite Scale
	if update.SpriteScale.IsSome() {
		vec := update.SpriteScale.Unwrap()
		if vec.X < MIN_SCALE_SIZE || vec.X > MAX_SCALE_SIZE ||
			vec.Y < MIN_SCALE_SIZE || vec.Y > MAX_SCALE_SIZE {
			update.SpriteScale = misc.NewOption(
				misc.Vector2{
					X: misc.ClampF64(vec.X, MIN_SCALE_SIZE, MAX_SCALE_SIZE),
					Y: misc.ClampF64(vec.Y, MIN_SCALE_SIZE, MAX_SCALE_SIZE),
				},
			)
		}
	}

	// Movement Speed
	if update.MovementSpeed.IsSome() {
		vec := update.MovementSpeed.Unwrap()
		if vec.X < MIN_MOVEMENT_SPEED || vec.X > MAX_MOVEMENT_SPEED {
			update.MovementSpeed = misc.NewOption(
				misc.Vector2{
					X: misc.ClampF64(vec.X, MIN_MOVEMENT_SPEED, MAX_MOVEMENT_SPEED),
					Y: 0,
				},
			)
		}
	}

	// Validate startPos
	if update.StartPos.IsSome() {
		vec := update.StartPos.Unwrap()
		if vec.X < 0 || vec.X > 1.0 || vec.Y < 0 || vec.Y > 1.0 {
			update.StartPos = misc.NewOption(
				misc.Vector2{
					X: misc.ClampF64(vec.X, 0, 1.0),
					Y: misc.ClampF64(vec.Y, 0, 1.0),
				},
			)
		}
	}

	// Validate actions
	if !IsActionEnum(update.CurrentAction) {
		update.CurrentAction = ACTION_PLAY_ANIMATION
		update.Action = NewActionPlayAnimation([]string{
			GetDefaultAnimForChibiStance(update.ChibiStance),
		})
	}
	if !update.Action.IsSet {
		update.CurrentAction = ACTION_PLAY_ANIMATION
		update.Action = NewActionPlayAnimation([]string{
			GetDefaultAnimForChibiStance(update.ChibiStance),
		})
	}

	// This fixes a bug where if the Opeator is in a walking action in base stance
	// and they change to a battle stance. They can't be in a walking action anymore
	// because operator battle chibis don't have move animations.
	// We need to explicitly set them to a play animation when then go into a
	// a battle stance.
	// Manually test this:
	// !chibi enemy b2 - should be in battle stance
	// !chibi walk     - makes the action wander
	// !chibi reed     - changes to reed. should be in base stance but still wander
	// !chibi battle   - changes to battle stance. should no longer be wandering
	if update.Faction == FACTION_ENUM_OPERATOR {
		if IsWalkingAction(update.CurrentAction) {
			if update.ChibiStance == CHIBI_STANCE_ENUM_BATTLE {
				update.CurrentAction = ACTION_PLAY_ANIMATION
				update.Action = NewActionPlayAnimation(
					[]string{GetDefaultAnimForChibiStance(update.ChibiStance)},
				)
			}
		}
	}

	switch update.CurrentAction {
	case ACTION_PLAY_ANIMATION:
		update.Action.Animations = getValidAnimations(
			update.AvailableAnimations,
			update.Action.Animations,
			GetDefaultAnimForChibiStance(update.ChibiStance),
		)
	case ACTION_WANDER:
		update.Action.WanderAnimation = getValidAnimations(
			update.AvailableAnimations,
			[]string{update.Action.WanderAnimation},
			GetDefaultAnimForChibiStance(update.ChibiStance),
		)[0]
	case ACTION_WALK_TO:
		if update.Action.TargetPos.IsNone() {
			update.Action.TargetPos = misc.NewOption(misc.Vector2{X: 0.5, Y: 0.5})
		} else {
			vec := update.Action.TargetPos.Unwrap()
			update.Action.TargetPos = misc.NewOption(
				misc.Vector2{
					X: misc.ClampF64(vec.X, 0, 1.0),
					Y: misc.ClampF64(vec.Y, 0, 1.0),
				},
			)
			availableAnimations := update.AvailableAnimations
			defaultAnimation := GetDefaultAnimForChibiStance(update.ChibiStance)
			update.Action.WalkToAnimation = getValidAnimations(
				availableAnimations,
				[]string{update.Action.WalkToAnimation},
				defaultAnimation,
			)[0]
			update.Action.WalkToFinalAnimation = getValidAnimations(
				availableAnimations,
				[]string{update.Action.WalkToFinalAnimation},
				defaultAnimation,
			)[0]
		}
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

func getValidAnimations(availableAnimations []string, actionAnimations []string, defaultAnim string) []string {
	for _, anim := range actionAnimations {
		if !slices.Contains(availableAnimations, anim) {
			actionAnimations = []string{defaultAnim}
			break
		}
	}
	// If it still doesn't exist then just choose one randomly
	for _, anim := range actionAnimations {
		if !slices.Contains(availableAnimations, anim) {
			actionAnimations = []string{availableAnimations[0]}
			break
		}
	}
	return actionAnimations
}
