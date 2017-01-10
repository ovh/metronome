package models

// BearerToken is the struct which is exposed by the /auth endpoint
type BearerToken struct {
	AccessToken  string `json:"token,omitempty"`
	Type         string `json:"tokenType"`
	RefreshToken string `json:"refreshToken,omitempty"`
}
