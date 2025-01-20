package spine

import "github.com/Stymphalian/ak_chibi_bot/server/internal/operator"

type FakeSpineClient struct {
	Users map[string]operator.OperatorInfo
}

func NewFakeSpineClient() *FakeSpineClient {
	return &FakeSpineClient{
		Users: make(map[string]operator.OperatorInfo, 0),
	}
}

func (f *FakeSpineClient) Close() error {
	return nil
}

func (f *FakeSpineClient) SetOperator(r *SetOperatorRequest) (*SetOperatorResponse, error) {
	f.Users[r.UserName] = r.Operator

	return &SetOperatorResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}

func (f *FakeSpineClient) RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error) {
	delete(f.Users, r.UserName)
	return &RemoveOperatorResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}

func (f *FakeSpineClient) ShowChatMessage(r *ShowChatMessageRequest) (*ShowChatMessageResponse, error) {
	return &ShowChatMessageResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   SHOW_CHAT_MESSAGE,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}

func (f *FakeSpineClient) FindOperator(r *FindOperatorRequest) (*FindOperatorResponse, error) {
	return &FindOperatorResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   FIND_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}
