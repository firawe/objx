package objx

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/stretchr/signature"
)

// JSON converts the contained object to a JSON string
// representation
func (o *O) JSON() (string, error) {

	result, err := json.Marshal(o.obj)

	if err != nil {
		err = errors.New("objx: JSON encode failed with: " + err.Error())
	}

	return string(result), err

}

// MustJSON converts the contained object to a JSON string
// representation and panics if there is an error
func (o *O) MustJSON() string {
	result, err := o.JSON()
	if err != nil {
		panic(err.Error())
	}
	return result
}

// Base64 converts the contained object to a Base64 string
// representation of the JSON string representation
func (o *O) Base64() (string, error) {

	var buf bytes.Buffer

	jsonData, err := o.JSON()
	if err != nil {
		return "", err
	}

	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(jsonData))
	encoder.Close()

	return buf.String(), nil

}

// MustBase64 converts the contained object to a Base64 string
// representation of the JSON string representation and panics
// if there is an error
func (o *O) MustBase64() string {
	result, err := o.Base64()
	if err != nil {
		panic(err.Error())
	}
	return result
}

// SignedBase64 converts the contained object to a Base64 string
// representation of the JSON string representation and signs it
// using the provided key.
func (o *O) SignedBase64(key string) (string, error) {

	base64, err := o.Base64()
	if err != nil {
		return "", err
	}

	sig := signature.HashWithKey([]byte(base64), []byte(key))

	return base64 + SignatureSeparator + sig, nil

}

// MustSignedBase64 converts the contained object to a Base64 string
// representation of the JSON string representation and signs it
// using the provided key and panics if there is an error
func (o *O) MustSignedBase64(key string) string {
	result, err := o.SignedBase64(key)
	if err != nil {
		panic(err.Error())
	}
	return result
}