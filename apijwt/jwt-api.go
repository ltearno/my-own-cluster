package apijwt

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ltearno/my-own-cluster/common"
	"github.com/ltearno/my-own-cluster/enginejs"
	"github.com/ltearno/my-own-cluster/enginewasm"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
)

type JWTAPIProvider struct {
	client        *http.Client
	trustProvider *TrustProviderConfiguration
	// issuer => kid => public key
	trustedKeys map[string]map[string]interface{}
}

type TrustProviderConfiguration struct {
	iss          string // what we match in the token's payload's iss
	certEndpoint string // endpoint url to fetch public keys
}

func NewJWTAPIProvider() (common.APIProvider, error) {
	p := &JWTAPIProvider{
		trustProvider: &TrustProviderConfiguration{
			iss:          "https://home.lteconsulting.fr",
			certEndpoint: "https://home.lteconsulting.fr/certs",
		},
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}},
		trustedKeys: make(map[string]map[string]interface{}),
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			set, err := jwk.Fetch(context.Background(), p.trustProvider.certEndpoint, jwk.WithHTTPClient(p.client))
			if err != nil {
				log.Printf("failed to parse JWK: %s %v", err, set)
			} else {
				trustedKeys := make(map[string]interface{})

				for it := set.Iterate(context.Background()); it.Next(context.Background()); {
					pair := it.Pair()
					key := pair.Value.(jwk.Key)

					var rawkey interface{} // This is the raw key, like *rsa.PrivateKey or *ecdsa.PrivateKey
					if err := key.Raw(&rawkey); err != nil {
						log.Printf("failed to create public key: %s", err)
					} else {
						trustedKeys[key.KeyID()] = rawkey

						exists := false
						existingKeys, ok := p.trustedKeys[p.trustProvider.iss]
						if ok {
							_, ok := existingKeys[key.KeyID()]
							exists = ok
						}
						if !exists {
							log.Printf("registered key:%s for trust:%s\n", key.KeyID(), p.trustProvider.certEndpoint)
						}
					}

				}

				p.trustedKeys[p.trustProvider.iss] = trustedKeys

				// TODO should clean up keys from this OIDC provider that are not published anymore
			}

			<-ticker.C
		}
	}()

	return p, nil
}

func (p *JWTAPIProvider) BindToExecutionEngineContext(ctx common.ExecutionEngineContextBounding) {
	wctx, ok := ctx.(*enginewasm.WasmProcessContext)
	if ok {
		wctx := *wctx
		BindJwtFunctionsWASM(wctx, p)
	}

	jsctx, ok := ctx.(*enginejs.JSProcessContext)
	if ok {
		BindJwtFunctionsJs(*jsctx, p)
	}
}

func VerifyJwt(ctx *common.FunctionExecutionContext, cookie interface{}, jwt string) (string, error) {
	r, e := vVerifyJwt(ctx, cookie, jwt)
	if e != nil {
		fmt.Printf("ERROR JWT %v\n", e)
	}
	return r, e
}

func vVerifyJwt(ctx *common.FunctionExecutionContext, cookie interface{}, jwt string) (string, error) {
	p := cookie.(*JWTAPIProvider)

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

	// check jwt's length

	// split jwt's three parts => 3 slices of the buffer
	protected, payload, signature, err := jws.SplitCompact(buf)
	if err != nil {
		return "", fmt.Errorf("bad formatted jwt")
	}

	// decode payload
	decodedPayload := make([]byte, base64.RawURLEncoding.DecodedLen(len(payload)))
	if _, err := base64.RawURLEncoding.Decode(decodedPayload, payload); err != nil {
		return "", fmt.Errorf(`failed to decode payload (%v)`, err)
	}

	// check we have issuer
	claims := make(map[string]interface{})
	err = json.Unmarshal(decodedPayload, &claims)
	if err != nil {
		return "", fmt.Errorf("undecodable payload")
	}
	issuerAny, ok := claims["iss"]
	if !ok {
		return "", fmt.Errorf("no issuer in claims")
	}
	issuer, ok := issuerAny.(string)
	if !ok {
		return "", fmt.Errorf("issuer not a string in claims (%v)", issuerAny)
	}
	keys, ok := p.trustedKeys[issuer]
	if !ok {
		return "", fmt.Errorf("no keys registered for issuer %s", issuer)
	}

	// decode header
	decodedHeader := make([]byte, base64.RawURLEncoding.DecodedLen(len(protected)))
	if _, err := base64.RawURLEncoding.Decode(decodedHeader, protected); err != nil {
		return "", fmt.Errorf(`failed to decode header (%v)`, err)
	}

	// check we have kid for issuer
	header := make(map[string]interface{})
	err = json.Unmarshal(decodedHeader, &header)
	if err != nil {
		return "", fmt.Errorf("undecodable header")
	}
	kidAny, ok := header["kid"]
	if !ok {
		return "", fmt.Errorf("no kid in header")
	}
	kid, ok := kidAny.(string)
	if !ok {
		return "", fmt.Errorf("kid is not a string in header (%v)", kidAny)
	}
	key, ok := keys[kid]
	if !ok {
		return "", fmt.Errorf("kid %s invalid for issuer %s", kid, issuer)
	}

	// check signature with info from the trust provider
	alg := jwa.RS256
	verifier, err := jws.NewVerifier(alg) // verify.New(alg)
	if err != nil {
		return "", fmt.Errorf(`failed to instantiate signature verifier %v`, err)
	}
	verifyBuf := &bytes.Buffer{}
	verifyBuf.Write(protected)
	verifyBuf.WriteByte('.')
	verifyBuf.Write(payload)
	decodedSignature := make([]byte, base64.RawURLEncoding.DecodedLen(len(signature)))
	if _, err := base64.RawURLEncoding.Decode(decodedSignature, signature); err != nil {
		return "", fmt.Errorf(`failed to decode signature %v`, err)
	}
	if err := verifier.Verify(verifyBuf.Bytes(), decodedSignature, key); err != nil {
		return "", fmt.Errorf(`failed to verify message (%v)`, err)
	}

	// return the token's body as a map
	return string(decodedPayload), nil
}
