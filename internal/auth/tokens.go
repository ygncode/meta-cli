package auth

import "time"

type Tokens struct {
	UserToken string               `json:"user_token"`
	ExpiresAt time.Time            `json:"expires_at,omitempty"`
	Pages     map[string]PageToken `json:"pages"`
}

type PageToken struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

func (t *Tokens) PageAccessToken(pageID string) (string, bool) {
	if t == nil || t.Pages == nil {
		return "", false
	}
	pt, ok := t.Pages[pageID]
	if !ok {
		return "", false
	}
	return pt.Token, true
}

func (t *Tokens) PageNames() map[string]string {
	m := make(map[string]string)
	for id, pt := range t.Pages {
		m[id] = pt.Name
	}
	return m
}
