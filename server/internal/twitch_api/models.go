package twitch_api

type GetUsersResponseData struct {
	Id              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageUrl string `json:"profile_image_url"`
	OfflineImageUrl string `json:"offline_image_url"`
	ViewCount       int    `json:"view_count"`
	Email           string `json:"email"`
	CreatedAt       string `json:"created_at"`
}

type GetUsersResponse struct {
	Data []GetUsersResponseData `json:"data"`
}

type StatusMessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RefreshTokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

type ValidateTokenResponse struct {
	ClientId string   `json:"client_id"`
	Login    string   `json:"login"`
	Scopes   []string `json:"scopes"`
	UserId   string   `json:"user_id"`
	Expires  int      `json:"expires_in"`
}

type GetOpenIdConfigurationResponse struct {
	//https://id.twitch.tv/oauth2/.well-known/openid-configuration
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	ClaimsParameterSupported          bool     `json:"claims_parameter_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	IdTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	Issuer                            string   `json:"issuer"`
	JwksUri                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	SubjectTypesSupported             []string `json:"suject_types_supported"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
}
