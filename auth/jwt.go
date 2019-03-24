package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alokic/gopkg/httputils"

	"github.com/pkg/errors"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/alokic/gopkg/contextutils"
	"github.com/alokic/gopkg/timeutils"
	"github.com/alokic/gopkg/typeutils"
)

var (
	// ErrTokenInvalid means token can not be verified with key.
	ErrTokenInvalid = errors.New("invalid jwt token")

	// ErrClaimsInvalid means claim can not be verified by user supplied handler.
	ErrClaimsInvalid = errors.New("invalid jwt claims")

	// ErrJWTKeyNotFound means jwt key is not set.
	ErrJWTKeyNotFound = errors.New("jwt key not found")

	// ErrUserNotAllowed means user's access is blocked.
	ErrUserNotAllowed = errors.New("user is not allowed")
)

// ClaimVerifier verifies claims from persistent store like database.
type ClaimVerifier func(map[string]interface{}) error

// JWT auth struct.
type JWT struct {
	key           string
	queryTokenKey string
	headerKey     string
	verifier      ClaimVerifier
	errHandler    http.Handler
}

// NewJWT is constructor of JWT.
func NewJWT(key string) *JWT {
	return &JWT{
		key:           key,
		queryTokenKey: "token",
		headerKey:     "Authorization",
		verifier: func(claims map[string]interface{}) error {
			return nil
		},
		errHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httputils.WriteError(ErrUserNotAllowed, http.StatusUnauthorized, w)
		}),
	}
}

// SetClaimVerifier set the handler to verify claim from DB.
func (j *JWT) SetClaimVerifier(v ClaimVerifier) {
	j.verifier = v
}

// SetErrHandler set the error handler used when jwt auth fail.
func (j *JWT) SetErrHandler(h http.Handler) {
	j.errHandler = h
}

// SetQueryTokenKey set the query token key.
func (j *JWT) SetQueryTokenKey(k string) {
	j.queryTokenKey = k
}

// SetHeaderKey set the header key which will contain the jwt token.
func (j *JWT) SetHeaderKey(k string) {
	j.headerKey = k
}

// Generate jwt token based claims passed.
func (j *JWT) Generate(claims map[string]interface{}) (string, error) {
	if j.key == "" {
		return "", errors.Wrap(ErrJWTKeyNotFound, "jwt.Generate")
	}
	// create the token
	token := jwtgo.New(jwtgo.SigningMethodHS256)

	newclaims := token.Claims.(jwtgo.MapClaims)

	// set claims
	for k, v := range claims {
		newclaims[k] = typeutils.ToStr(v)
	}

	newclaims["_ts"] = fmt.Sprint(timeutils.UnixTime())

	s, e := token.SignedString([]byte(j.key))
	if e != nil {
		return "", errors.Wrap(e, "jwt.Generate")
	}
	return s, nil
}

// VerifyToken verifies jwt tokens.
func (j *JWT) VerifyToken(jwttoken string) (map[string]interface{}, error) {
	token, err := jwtgo.Parse(jwttoken, func(token *jwtgo.Token) (interface{}, error) {
		return []byte(j.key), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "jwt.Verify")
	}

	if !token.Valid {
		return nil, errors.Wrap(ErrTokenInvalid, "jwt.Verify")
	}

	return token.Claims.(jwtgo.MapClaims), nil
}

// VerifyClaims verifies claims form user supplied handler.
func (j *JWT) VerifyClaims(claims map[string]interface{}) error {
	err := j.verifier(claims)
	if err != nil {
		return errors.Wrap(ErrClaimsInvalid, "jwt.VerifyClaims")
	}
	return nil
}

// HTTPMiddleware wraps a http.Handler to perform jwt auth on request.
func (j *JWT) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get(j.queryTokenKey)
		if token == "" {
			token = j.headerToken(r)
		}

		claims, err := j.VerifyToken(token)
		if err != nil {
			j.errHandler.ServeHTTP(w, r)
			return
		}

		err = j.VerifyClaims(claims)
		if err != nil {
			j.errHandler.ServeHTTP(w, r)
			return
		}

		ctx := contextutils.SetParams(r.Context(), claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (j *JWT) headerToken(r *http.Request) string {
	if ah := r.Header.Get(j.headerKey); ah != "" {
		if len(ah) > 6 && strings.ToUpper(ah[0:7]) == "BEARER " {
			return ah[7:]
		}
	}
	return ""
}
