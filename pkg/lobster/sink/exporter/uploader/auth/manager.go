/*
 * Copyright (c) 2024-present NAVER Corp
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang/glog"
	cache "github.com/hashicorp/golang-lru"

	v1 "github.com/naver/lobster/pkg/operator/api/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type TokenManager struct {
	sync.RWMutex
	cache *cache.Cache
}

func NewTokenManager() *TokenManager {
	c, _ := cache.New(100)
	return &TokenManager{sync.RWMutex{}, c}
}

func (tm *TokenManager) GetOAuthTokenProvider(ctx context.Context, oAuthType v1.OAuthType, conf *clientcredentials.Config) (TokenProvider, error) {
	key := generateKey(fmt.Sprintf("%s:%s:%s", conf.TokenURL, conf.ClientID, conf.ClientSecret))

	tm.RLock()
	v, ok := tm.cache.Get(key)
	tm.RUnlock()
	if ok && v.(*oauth2.Token).Valid() {
		glog.Info("reuse " + v.(*oauth2.Token).AccessToken)
		return TokenProvider{oauth2.StaticTokenSource(v.(*oauth2.Token))}, nil
	}

	rtt, err := NewAuthRoundTripper(oAuthType, conf)
	if err != nil {
		return TokenProvider{}, err
	}

	newToken, err := conf.TokenSource(
		context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: rtt}),
	).Token()
	if err != nil {
		return TokenProvider{}, err
	}

	glog.Info("new " + newToken.AccessToken)

	tm.Lock()
	tm.cache.Add(key, newToken)
	tm.Unlock()

	return TokenProvider{oauth2.StaticTokenSource(newToken)}, nil
}

func generateKey(str string) string {
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}
