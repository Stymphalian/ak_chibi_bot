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
