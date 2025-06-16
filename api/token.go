package api

type TokenMerchantBinding struct {
	BindingType  string `json:"bindingType"`
	MdxID        string `json:"mdxId"`
	MerchantName any    `json:"merchantName"`
	URLID        string `json:"urlId"`
}

type TokenRules struct {
	AllowAuthorizations bool                 `json:"allowAuthorizations"`
	MerchantBinding     TokenMerchantBinding `json:"merchantBinding"`
}

type Token struct {
	Token            string     `json:"token"`
	Cvv              string     `json:"cvv"`
	ExpirationDate   string     `json:"expirationDate"`
	LastFour         string     `json:"lastFour"`
	TokenReferenceID string     `json:"tokenReferenceId"`
	CreatedTimestamp string     `json:"createdTimestamp"`
	TokenName        string     `json:"tokenName"`
	TokenStatus      string     `json:"tokenStatus"`
	TokenType        string     `json:"tokenType"`
	TokenRules       TokenRules `json:"tokenRules"`
	CardReferenceID  string     `json:"cardReferenceId"`
}
