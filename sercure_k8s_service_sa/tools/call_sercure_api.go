package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/itdevsamurai/gke/sercure_k8s_service_sa/tools/api/config"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
)

func main() {
	jwt, err := generateJWT(config.ServiceAccountFile, config.ServiceAccountEmail, config.Audience, 3600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\njwt:\n%s\n\n", jwt)

	resp, err := makeJWTRequest(jwt, config.ApiUrl)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s Response: %s", config.ApiUrl, resp)
}

// generateJWT creates a signed JSON Web Token using a Google API Service Account.
func generateJWT(saKeyfile, saEmail, audience string, expiryLength int64) (string, error) {
	now := time.Now().Unix()

	// Build the JWT payload.
	jwt := &jws.ClaimSet{
		Iat: now,
		// expires after 'expiraryLength' seconds.
		Exp: now + expiryLength,
		// Iss must match 'issuer' in the security configuration in your
		// swagger spec (e.g. service account email). It can be any string.
		Iss: saEmail,
		// Aud must be either your Endpoints service name, or match the value
		// specified as the 'x-google-audience' in the OpenAPI document.
		Aud: audience,
		// Sub and Email should match the service account's email address.
		Sub:           saEmail,
		PrivateClaims: map[string]interface{}{"email": saEmail},
	}
	jwsHeader := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
	}

	// Extract the RSA private key from the service account keyfile.
	sa, err := ioutil.ReadFile(saKeyfile)
	if err != nil {
		return "", fmt.Errorf("Could not read service account file: %v", err)
	}
	conf, err := google.JWTConfigFromJSON(sa)
	if err != nil {
		return "", fmt.Errorf("Could not parse service account JSON: %v", err)
	}
	block, _ := pem.Decode(conf.PrivateKey)
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("private key parse error: %v", err)
	}
	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	// Sign the JWT with the service account's private key.
	if !ok {
		return "", errors.New("private key failed rsa.PrivateKey type assertion")
	}
	return jws.Encode(jwsHeader, jwt, rsaKey)
}

// makeJWTRequest sends an authorized request to your deployed endpoint.
func makeJWTRequest(signedJWT, url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	reqBody, err := json.Marshal(config.RequestData)

	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request body: %v", err)
	}

	req, err := http.NewRequest(config.RequestMethod, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+signedJWT)
	req.Header.Add("content-type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTTP response: %v", err)
	}
	return string(responseData), nil
}
