package web_test

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func unsignedIDToken(payload string) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"none"}`)) + "." + enc([]byte(payload)) + "."
}

func TestResolveOAuthIdentityRequiresVerificationOutsideDev(t *testing.T) {
	prev := config.AppConfig.DevMode
	config.AppConfig.DevMode = false
	t.Cleanup(func() { config.AppConfig.DevMode = prev })

	tok := unsignedIDToken(`{"sub":"s","email":"a@b.c","email_verified":true}`)
	_, _, _, err := web.ResolveOAuthIdentityForTest(context.Background(),
		map[string]interface{}{"provider_type": "oidc"},
		map[string]interface{}{"id_token": tok},
	)
	if err == nil || !strings.Contains(err.Error(), "verification required") {
		t.Fatalf("err %v", err)
	}
}

func TestResolveOAuthIdentityEmailVerifiedFlag(t *testing.T) {
	prev := config.AppConfig.DevMode
	config.AppConfig.DevMode = true
	t.Cleanup(func() { config.AppConfig.DevMode = prev })

	prov := map[string]interface{}{"provider_type": "oidc"}
	verifiedTok := unsignedIDToken(`{"sub":"s","email":"a@b.c","email_verified":true}`)
	email, sub, verified, err := web.ResolveOAuthIdentityForTest(context.Background(), prov, map[string]interface{}{"id_token": verifiedTok})
	if err != nil || !verified || email != "a@b.c" || sub != "s" {
		t.Fatalf("email=%q sub=%q verified=%v err=%v", email, sub, verified, err)
	}

	plainTok := unsignedIDToken(`{"sub":"s","email":"a@b.c"}`)
	_, _, verified, err = web.ResolveOAuthIdentityForTest(context.Background(), prov, map[string]interface{}{"id_token": plainTok})
	if err != nil || verified {
		t.Fatalf("verified=%v err=%v", verified, err)
	}
}
