package token

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "errors"
    "strings"
    "time"
)

type Claims struct {
    UserID uint  `json:"uid"`
    Exp    int64 `json:"exp"`
}

type header struct {
    Alg string `json:"alg"`
    Typ string `json:"typ"`
}

func Generate(userID uint, ttl time.Duration, secret string) (string, time.Time, error) {
    if secret == "" {
        return "", time.Time{}, errors.New("token secret is empty")
    }

    exp := time.Now().Add(ttl)
    h := header{Alg: "HS256", Typ: "JWT"}
    p := Claims{UserID: userID, Exp: exp.Unix()}

    hb, err := json.Marshal(h)
    if err != nil {
        return "", time.Time{}, err
    }
    pb, err := json.Marshal(p)
    if err != nil {
        return "", time.Time{}, err
    }

    enc := base64.RawURLEncoding
    hp := enc.EncodeToString(hb) + "." + enc.EncodeToString(pb)
    sig := sign(hp, secret)

    return hp + "." + sig, exp, nil
}

func Validate(tokenString, secret string) (Claims, error) {
    if secret == "" {
        return Claims{}, errors.New("token secret is empty")
    }

    parts := strings.Split(tokenString, ".")
    if len(parts) != 3 {
        return Claims{}, errors.New("invalid token format")
    }

    hp := parts[0] + "." + parts[1]
    expected := sign(hp, secret)
    if !hmac.Equal([]byte(expected), []byte(parts[2])) {
        return Claims{}, errors.New("invalid token signature")
    }

    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return Claims{}, errors.New("invalid token payload")
    }

    var claims Claims
    if err := json.Unmarshal(payload, &claims); err != nil {
        return Claims{}, errors.New("invalid token claims")
    }

    if time.Now().Unix() > claims.Exp {
        return Claims{}, errors.New("token expired")
    }

    return claims, nil
}

func sign(data, secret string) string {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(data))
    return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}