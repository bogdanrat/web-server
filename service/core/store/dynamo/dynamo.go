package dynamo

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/bogdanrat/web-server/service/core/store"
	"log"
	"os"
	"strings"
	"time"
)

const (
	tableStatusActive     = "ACTIVE"
	tableStatusMaxRetries = 3
)

type KeyValueStore struct {
	svc       *dynamodb.DynamoDB
	tableName string
}

func NewStore(tableName string) (store.KeyValue, error) {
	creds, err := lib.GetRoleCredentials(config.AppConfig.AWS.DynamoDBRoleARN, os.Getenv("DYNAMODB_ROLE_EXTERNAL_ID"))
	if err != nil {
		return nil, err
	}

	keyValueStore := &KeyValueStore{
		svc:       dynamodb.New(config.AWSSession, &aws.Config{Credentials: creds}),
		tableName: tableName,
	}

	if err = keyValueStore.setup(); err != nil {
		return nil, err
	}

	err = keyValueStore.Put(&models.KeyValuePair{
		Key:   "test",
		Value: "message",
	})
	if err != nil {
		log.Println(err)
	}

	return keyValueStore, nil
}

func (s *KeyValueStore) Get(key string) (interface{}, error) {
	output, err := s.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			store.KeyIdentifier: {
				S: aws.String(key),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var value interface{}
	err = dynamodbattribute.Unmarshal(output.Item[store.ValueIdentifier], &value)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, store.KeyNotFoundError
	}
	return value, nil
}

func (s *KeyValueStore) Put(pair *models.KeyValuePair) error {
	attributeValues, err := dynamodbattribute.MarshalMap(pair)
	if err != nil {
		return err
	}

	_, err = s.svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      attributeValues,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *KeyValueStore) Delete(key string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			store.KeyIdentifier: {
				S: aws.String(key),
			},
		},
		ReturnValues: aws.String("NONE"),
	}

	_, err := s.svc.DeleteItem(input)
	if err != nil {
		return err
	}

	return nil
}

func (s *KeyValueStore) GetAll() ([]*models.KeyValuePair, error) {
	projection := expression.NamesList(expression.Name(store.KeyIdentifier), expression.Name(store.ValueIdentifier))
	expr, err := expression.NewBuilder().WithProjection(projection).Build()
	if err != nil {
		return nil, fmt.Errorf("error building expression: %s", err)
	}

	queryParams := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(s.tableName),
	}

	result, err := s.svc.Scan(queryParams)
	if err != nil {
		return nil, fmt.Errorf("error scaning table: %s", err)
	}

	keyValuePairs := make([]*models.KeyValuePair, 0)

	for _, item := range result.Items {
		pair := &models.KeyValuePair{}

		err = dynamodbattribute.UnmarshalMap(item, pair)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling table item: %s", err)
		}

		keyValuePairs = append(keyValuePairs, pair)
	}

	return keyValuePairs, nil
}

func (s *KeyValueStore) PutMany(pairs []*models.KeyValuePair) error {
	return s.writeBatch(pairs)
}

func (s *KeyValueStore) setup() error {
	exists, err := s.tableExists(s.tableName)
	if err != nil {
		return err
	}

	if !exists {
		tableDescription, err := s.createTable(s.tableName)
		if err != nil {
			return err
		}
		log.Printf("Created I18N Table: %s\n", *tableDescription.TableArn)

		if config.AppConfig.I18N.Seed {
			err = s.seed()
			if err != nil {
				log.Println(err)
			}
			log.Println("Seeded I18N Table.")
		}
	}

	log.Println("I18N initialized.")
	return nil
}

func (s *KeyValueStore) tableExists(tableName string) (bool, error) {
	found := false
	listTablesInput := &dynamodb.ListTablesInput{}

	for {
		listTablesOutput, err := s.svc.ListTables(listTablesInput)
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

func (s *KeyValueStore) createTable(tableName string) (*dynamodb.TableDescription, error) {
	createTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(store.KeyIdentifier),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(store.KeyIdentifier),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(config.AppConfig.AWS.DynamoDBRCU),
			WriteCapacityUnits: aws.Int64(config.AppConfig.AWS.DynamoDBWCU),
		},
	}

	request, response := s.svc.CreateTableRequest(createTableInput)
	if err := request.Send(); err != nil {
		return nil, err
	}

	return response.TableDescription, nil
}

func (s *KeyValueStore) seed() error {
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
				active, routineErr = s.checkTableStatus(s.tableName, tableStatusActive)
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
		return routineErr
	}
	if !active {
		return fmt.Errorf("table did not become active in due time")
	}

	log.Println("Seeding table...")

	err := s.writeBatch(config.AppConfig.I18N.SeedValues)
	if err != nil {
		return err
	}
	return nil
}

func (s *KeyValueStore) writeBatch(batch []*models.KeyValuePair) error {
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: make(map[string][]*dynamodb.WriteRequest),
	}

	for _, item := range batch {
		input.RequestItems[s.tableName] = append(input.RequestItems[s.tableName],
			&dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: map[string]*dynamodb.AttributeValue{
						store.KeyIdentifier: {
							S: aws.String(item.Key),
						},
						store.ValueIdentifier: {
							S: aws.String(item.Value.(string)),
						},
					},
				},
			},
		)
	}

	_, err := s.svc.BatchWriteItem(input)
	if err != nil {
		return err
	}
	return nil
}

// checkTableStatus checks whether a given table has a given status (CREATING, ACTIVE...)
func (s *KeyValueStore) checkTableStatus(tableName, status string) (bool, error) {
	describeOutput, err := s.svc.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
	if err != nil {
		return false, err
	}

	if describeOutput.Table.TableStatus != nil && strings.EqualFold(*describeOutput.Table.TableStatus, status) {
		return true, nil
	}
	return false, nil
}
