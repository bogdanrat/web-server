package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/bogdanrat/web-server/service/core/common"
	"github.com/bogdanrat/web-server/service/core/config"
)

var (
	secretsService *secretsmanager.SecretsManager
)

func GetDatabaseSecrets() (*common.DatabaseSecrets, error) {
	if secretsService == nil {
		secretsService = secretsmanager.New(config.AWSSession)
	}

	secretName := config.AppConfig.AWS.DatabaseSecretARN

	result, err := secretsService.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve secret %s", secretName)
	}

	secrets := &common.DatabaseSecrets{}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secretString, decodedBinarySecret string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		secretLen, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			return nil, fmt.Errorf("base64 decode rrror: %s", err)
		}
		decodedBinarySecret = string(decodedBinarySecretBytes[:secretLen])
	}

	if secretString != "" {
		err = json.Unmarshal([]byte(secretString), secrets)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal database secrets: %s", err)
		}
	} else {
		err = json.Unmarshal([]byte(decodedBinarySecret), secrets)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal database secrets: %s", err)
		}
	}

	return secrets, nil
}
