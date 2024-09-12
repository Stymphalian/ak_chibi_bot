package spine

func NewTestAssetService() *AssetService {
	a := &AssetService{
		AssetMap:         NewSpineAssetMap(),
		CommonNames:      NewCommonNames(),
		EnemyAssetMap:    NewSpineAssetMap(),
		EnemyCommonNames: NewCommonNames(),
	}
	a.AssetMap.Data["char_002_amiya"] = &ChibiAssetPathEntry{
		Skins: map[string]*SpineSkinData{
			DEFAULT_SKIN_NAME: {
				Base: map[ChibiFacingEnum]*SpineData{
					CHIBI_FACING_ENUM_FRONT: {
						AtlasFilepath: "base_front_atlas_filepath",
						SkelFilepath:  "base_front_skel_filepath",
						PngFilepath:   "base_front_png_filepath",
						Animations:    []string{DEFAULT_ANIM_BASE, "base_front1", "base_front2"},
					},
				},
				Battle: map[ChibiFacingEnum]*SpineData{
					CHIBI_FACING_ENUM_FRONT: {
						AtlasFilepath: "battle_front_atlas_filepath",
						SkelFilepath:  "battle_front_skel_filepath",
						PngFilepath:   "battle_front_png_filepath",
						Animations:    []string{DEFAULT_ANIM_BATTLE, "battle_front1", "battle_front2"},
					},
					CHIBI_FACING_ENUM_BACK: {
						AtlasFilepath: "battle_back_atlas_filepath",
						SkelFilepath:  "battle_back_skel_filepath",
						PngFilepath:   "battle_back_png_filepath",
						Animations:    []string{DEFAULT_ANIM_BATTLE, "battle_back1", "battle_back2"},
					},
				},
			},
		},
	}
	a.CommonNames.allNames = []string{"Amiya"}
	a.CommonNames.operatorIdToNames = map[string][]string{"char_002_amiya": {"Amiya"}}
	a.CommonNames.namesToOperatorId = map[string][]string{"Amiya": {"char_002_amiya"}}

	a.EnemyAssetMap.Data["enemy_1007_slime_2"] = &ChibiAssetPathEntry{
		Skins: map[string]*SpineSkinData{
			DEFAULT_SKIN_NAME: {
				Battle: map[ChibiFacingEnum]*SpineData{
					CHIBI_FACING_ENUM_FRONT: {
						AtlasFilepath: "battle_front_atlas_filepath",
						SkelFilepath:  "battle_front_skel_filepath",
						PngFilepath:   "battle_front_png_filepath",
						Animations:    []string{DEFAULT_ANIM_BATTLE, "battle_front1", "battle_front2", "move"},
					},
				},
			},
		},
	}
	a.EnemyCommonNames.allNames = []string{"Slug"}
	a.EnemyCommonNames.operatorIdToNames = map[string][]string{"enemy_1007_slime_2": {"Slug"}}
	a.EnemyCommonNames.namesToOperatorId = map[string][]string{"Slug": {"enemy_1007_slime_2"}}
	return a
}
