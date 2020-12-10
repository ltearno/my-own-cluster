package apijwt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"my-own-cluster/common"
	"my-own-cluster/enginejs"
	"my-own-cluster/enginewasm"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
)

type JWTAPIProvider struct {
	client *http.Client
}

func NewJWTAPIProvider() (common.APIProvider, error) {
	p := &JWTAPIProvider{
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}},
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				fmt.Printf("TICKER\n")
			}
		}
	}()

	return p, nil
}

func (p *JWTAPIProvider) BindToExecutionEngineContext(ctx common.ExecutionEngineContextBounding) {
	wctx, ok := ctx.(*enginewasm.WasmProcessContext)
	if ok {
		wctx := *wctx
		BindJwtFunctionsWASM(wctx)
	}

	jsctx, ok := ctx.(*enginejs.JSProcessContext)
	if ok {
		BindJwtFunctionsJs(*jsctx)
	}
}

// should be elsewhere but...
func VerifyJwt(ctx *common.FunctionExecutionContext, jwt string) (string, error) {
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	set, err := jwk.FetchHTTP("https://192.168.0.2:8443/certs", jwk.WithHTTPClient(client))
	if err != nil {
		log.Printf("failed to parse JWK: %s %v", err, set)
		return "", nil
	}

	// sample payload
	/*
		{
			"alg": "RS256",
			"kid": "N0IabvUCXr4UyWeNlphOFVl8U9edF7fKt6oiVWilVfRl8B7PLIsF0uUgWbH4w-_pC6gV_i78yunsNKz82_tnDg",
			"typ": "JWT"
		}
		{
			"iss": "https://home.lteconsulting.fr",
			"sub": "LTE Consulting",
			"aud": "IDP",
			"exp": 1607628782,
			"iat": 1607614382,
			"jti": "78db8360-6be9-4bcc-a14b-0721528f96bd",
			"uuid": "ltearno",
			"role": "{}",
			"roles": "{}"
		}
	*/

	buf := []byte(jwt)
	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		var rawkey interface{} // This is the raw key, like *rsa.PrivateKey or *ecdsa.PrivateKey
		if err := key.Raw(&rawkey); err != nil {
			log.Printf("failed to create public key: %s", err)
			return "", err
		}

		verified, err := jws.Verify(buf, jwa.RS256, rawkey)
		if err != nil {
			log.Printf("failed to verify message: %s", err)
			return "", err
		}

		log.Printf("verified: %v", string(verified))
	}
	log.Fatal("bye")
	return "", nil
}
