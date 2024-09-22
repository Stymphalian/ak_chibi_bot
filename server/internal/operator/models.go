package operator

// import "github.com/Stymphalian/ak_chibi_bot/server/internal/operator"

type AnimationsList []string
type FacingData struct {
	Facings map[ChibiFacingEnum]AnimationsList `json:"facing"`
}

func (f *FacingData) HasFacing(facing ChibiFacingEnum) bool {
	if _, ok := f.Facings[facing]; !ok {
		return false
	}
	animations := f.Facings[facing]
	return len(animations) != 0
}

type SkinData struct {
	Stances map[ChibiStanceEnum]FacingData `json:"stance"`
}

func (s *SkinData) HasChibiStance(chibiStance ChibiStanceEnum) bool {
	if _, ok := s.Stances[chibiStance]; !ok {
		return false
	}
	faceData := s.Stances[chibiStance]
	return len(faceData.Facings) != 0
}

type GetOperatorResponse struct {
	OperatorId   string              `json:"operator_id"` // char_002_amiya
	OperatorName string              `json:"operator_name"`
	Skins        map[string]SkinData `json:"skins"` // build_char_002_amiya
}

func (r *GetOperatorResponse) GetSkinNames() []string {
	skins := make([]string, len(r.Skins))
	i := 0
	for skinName := range r.Skins {
		skins[i] = skinName
		i += 1
	}
	return skins
}
