//go:build integration

package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"sumeru/core/orm"
	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

const rpcAuthTestPasswordHash = "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // "password"

func rpcAuthIntegrationDB(t *testing.T) context.Context {
	t.Helper()
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	preflight, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := preflight.Ping(); err != nil {
		_ = preflight.Close()
		t.Fatalf("db ping: %v", err)
	}
	_ = preflight.Close()

	orm.InitDBWithPool(dsn, orm.DBPoolSettings{MaxOpenConns: 5, MaxIdleConns: 2})
	if !orm.IsInitialized() {
		t.Skip("database schema not bootstrapped (run sumeru -i base)")
	}
	if config.AppConfig.DbName == "" {
		config.AppConfig.DbName = "sumeru"
	}
	ctx := orm.ContextWithBypass(context.Background(), true)
	return orm.ContextWithUID(ctx, 1)
}

func TestRPCAuthLogin_successAndBearerRPC(t *testing.T) {
	ctx := rpcAuthIntegrationDB(t)
	userModel, ok := orm.Registry["core.user"]
	if !ok {
		t.Fatal("core.user not registered")
	}
	login := fmt.Sprintf("rpc_auth_%d@test.local", time.Now().UnixNano())
	userID, err := orm.Create(ctx, userModel, map[string]interface{}{
		"login": login,
		"name":  "RPC Auth Test",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = orm.Unlink(ctx, userModel, []int{userID})
	})
	if err := orm.SetUserPasswordHash(ctx, userID, rpcAuthTestPasswordHash); err != nil {
		t.Fatalf("SetUserPasswordHash: %v", err)
	}

	web.ResetLoginLockoutForTest()
	t.Cleanup(web.ResetLoginLockoutForTest)

	loginBody := fmt.Sprintf(`{"login":%q,"password":"password"}`, login)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, web.TestAuthLoginURLForTest(), strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	web.ApplyDefaultRPCHeadersForTest(req)
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			UID       int    `json:"uid"`
			APIKey    string `json:"api_key"`
			SessionID string `json:"session_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !env.Success || env.Data.UID != userID || env.Data.APIKey == "" || !strings.HasPrefix(env.Data.APIKey, "sk_") {
		t.Fatalf("bad success payload: %+v", env)
	}
	if env.Data.SessionID == "" {
		t.Fatal("expected session_id")
	}

	rpcRR := httptest.NewRecorder()
	rpcReq := httptest.NewRequest(http.MethodPost, web.TestRPCURLForTest("core.user", "search"), strings.NewReader(`{"domain":[["id","=",`+fmt.Sprint(userID)+`]],"fields":["login"],"limit":1}`))
	rpcReq.Header.Set("Content-Type", "application/json")
	rpcReq.Header.Set("Authorization", "Bearer "+env.Data.APIKey)
	web.ApplyDefaultRPCHeadersForTest(rpcReq)
	web.RPCJSONHandler(rpcRR, rpcReq)
	if rpcRR.Code != http.StatusOK {
		t.Fatalf("Bearer RPC status = %d body = %s", rpcRR.Code, rpcRR.Body.String())
	}
	rpcEnv := decodeRPCEnvelope(t, rpcRR)
	if !rpcEnv.Success {
		t.Fatalf("RPC success = false: %+v", rpcEnv.Error)
	}
}

func TestAuthenticateLocalUser_integration(t *testing.T) {
	ctx := rpcAuthIntegrationDB(t)
	userModel := orm.Registry["core.user"]
	login := fmt.Sprintf("orm_auth_%d@test.local", time.Now().UnixNano())
	userID, err := orm.Create(ctx, userModel, map[string]interface{}{
		"login": login,
		"name":  "ORM Auth Test",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = orm.Unlink(ctx, userModel, []int{userID})
	})
	if err := orm.SetUserPasswordHash(ctx, userID, rpcAuthTestPasswordHash); err != nil {
		t.Fatalf("SetUserPasswordHash: %v", err)
	}
	uid, username, name, err := orm.AuthenticateLocalUser(ctx, login, "password")
	if err != nil || uid != userID || username != login || name == "" {
		t.Fatalf("AuthenticateLocalUser = uid=%d name=%q err=%v", uid, name, err)
	}
	_, _, _, err = orm.AuthenticateLocalUser(ctx, login, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}
