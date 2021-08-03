package dynamo

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/i18n"
	"log"
	"os"
	"strings"
	"time"
)

const (
	tableStatusActive     = "ACTIVE"
	tableStatusMaxRetries = 3
)

type Translator struct {
	svc           *dynamodb.DynamoDB
	keyValuePairs map[string]string
}

func NewTranslator() (i18n.Translator, error) {
	dynamoService, err := setup()
	if err != nil {
		return nil, err
	}

	translator := &Translator{
		svc:           dynamoService,
		keyValuePairs: make(map[string]string),
	}

	exists, err := translator.tableExists(config.AppConfig.I18N.TableName)
	if err != nil {
		return nil, err
	}

	if !exists {
		tableDescription, err := translator.createTable(config.AppConfig.I18N.TableName)
		if err != nil {
			return nil, err
		}
		log.Printf("Created I18N Table: %s\n", *tableDescription.TableArn)

		if config.AppConfig.I18N.Seed {
			_, err := translator.seed(config.AppConfig.I18N.TableName)
			if err != nil {
				log.Println(err)
			}
			log.Println("Seeded I18N Table.")
		}
	} else {
		if err = translator.Reload(); err != nil {
			log.Printf("Could not load translations: %v\n", err)
		} else {
			log.Println("I18N initialized.")
		}
	}

	return translator, nil
}

func (t *Translator) Do(key string, substitutions map[string]string) string {
	translation, ok := t.keyValuePairs[key]
	if ok {
		for find, replace := range substitutions {
			translation = strings.Replace(translation, "{{"+find+"}}", replace, -1)
		}
	}
	return translation
}

func setup() (*dynamodb.DynamoDB, error) {
	creds, err := getCredentials()
	if err != nil {
		return nil, err
	}
	return dynamodb.New(config.AWSSession, &aws.Config{Credentials: creds}), nil
}

func getCredentials() (*credentials.Credentials, error) {
	creds := stscreds.NewCredentials(config.AWSSession, config.AppConfig.AWS.DynamoDBRoleARN, func(provider *stscreds.AssumeRoleProvider) {
		provider.ExternalID = aws.String(os.Getenv("DYNAMODB_ROLE_EXTERNAL_ID"))
	})

	return creds, nil
}

func (t *Translator) tableExists(tableName string) (bool, error) {
	found := false
	listTablesInput := &dynamodb.ListTablesInput{}

	for {
		listTablesOutput, err := t.svc.ListTables(listTablesInput)
		if err != nil {
			return false, err
		}

		for _, table := range listTablesOutput.TableNames {
			if table != nil && *table == tableName {
				found = true
				break
			}
		}

		listTablesInput.ExclusiveStartTableName = listTablesOutput.LastEvaluatedTableName
		if listTablesOutput.LastEvaluatedTableName == nil {
			break
		}
	}

	return found, nil
}

func (t *Translator) createTable(tableName string) (*dynamodb.TableDescription, error) {
	createTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Key"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Key"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(config.AppConfig.AWS.DynamoDBRCU),
			WriteCapacityUnits: aws.Int64(config.AppConfig.AWS.DynamoDBWCU),
		},
	}

	request, response := t.svc.CreateTableRequest(createTableInput)
	if err := request.Send(); err != nil {
		return nil, err
	}

	return response.TableDescription, nil
}

func (t *Translator) seed(tableName string) (*dynamodb.BatchWriteItemOutput, error) {
	ticker := time.NewTicker(2 * time.Second)
	waitChan := make(chan bool)
	var active bool
	var routineErr error
	var retries = 0

	go func() {
		defer ticker.Stop()
		defer close(waitChan)

		for {
			select {
			case <-ticker.C:
				active, routineErr = t.checkTableStatus(tableName, tableStatusActive)
				if routineErr != nil || active {
					return
				}
				retries++
				if retries >= tableStatusMaxRetries {
					return
				}
			}
		}
	}()

	log.Println("Waiting for table to become active...")
	<-waitChan
	if routineErr != nil {
		return nil, routineErr
	}
	if !active {
		return nil, fmt.Errorf("table did not become active in due time")
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: make(map[string][]*dynamodb.WriteRequest),
	}

	for _, seedValue := range config.AppConfig.I18N.SeedValues {
		attributeValues, err := dynamodbattribute.MarshalMap(seedValue)
		if err != nil {
			return nil, err
		}

		input.RequestItems[tableName] = append(input.RequestItems[tableName],
			&dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: attributeValues,
				},
			},
		)
	}

	log.Println("Seeding table...")
	result, err := t.svc.BatchWriteItem(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// checkTableStatus checks whether a given table has a given status (CREATING, ACTIVE...)
func (t *Translator) checkTableStatus(tableName, status string) (bool, error) {
	describeOutput, err := t.svc.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
	if err != nil {
		return false, err
	}

	if describeOutput.Table.TableStatus != nil && strings.EqualFold(*describeOutput.Table.TableStatus, status) {
		return true, nil
	}
	return false, nil
}

func (t *Translator) Reload() error {
	projection := expression.NamesList(expression.Name("Key"), expression.Name("Value"))
	expr, err := expression.NewBuilder().WithProjection(projection).Build()
	if err != nil {
		return fmt.Errorf("error building expression: %s", err)
	}

	queryParams := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(config.AppConfig.I18N.TableName),
	}

	result, err := t.svc.Scan(queryParams)
	if err != nil {
		return fmt.Errorf("error scaning table: %s", err)
	}

	keyValuePairs := make([]*config.KeyValuePair, 0)

	for _, item := range result.Items {
		pair := &config.KeyValuePair{}

		err = dynamodbattribute.UnmarshalMap(item, pair)
		if err != nil {
			return fmt.Errorf("error unmarshalling table item: %s", err)
		}

		keyValuePairs = append(keyValuePairs, pair)
	}

	t.keyValuePairs = make(map[string]string)
	for _, pair := range keyValuePairs {
		t.keyValuePairs[pair.Key] = pair.Value
	}

	return nil
}
