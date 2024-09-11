package spine

type FakeSpineClient struct {
	Users map[string]OperatorInfo
}

func NewFakeSpineClient() *FakeSpineClient {
	return &FakeSpineClient{
		Users: make(map[string]OperatorInfo, 0),
	}
}

func (f *FakeSpineClient) Close() error {
	return nil
}

func (f *FakeSpineClient) SetOperator(r *SetOperatorRequest) (*SetOperatorResponse, error) {
	f.Users[r.UserName] = r.Operator

	return &SetOperatorResponse{
		SpineResponse: SpineResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}

func (f *FakeSpineClient) RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error) {
	delete(f.Users, r.UserName)
	return &RemoveOperatorResponse{
		SpineResponse: SpineResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}

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
	a.CommonNames.allNames = []string{"Amiya"}
	a.CommonNames.operatorIdToNames = map[string][]string{"char_002_amiya": {"Amiya"}}
	a.CommonNames.namesToOperatorId = map[string][]string{"Amiya": {"char_002_amiya"}}
	return a
}
