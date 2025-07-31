package dynamo

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func GetSession() *session.Session {
	env := os.Getenv("ENVIRONMENT")
	if env == "CLOUD" {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		fmt.Println("Iniciando Sessão Cloud")
		return sess
	} else {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region:   aws.String("us-east-1"),
				Endpoint: aws.String("http://dynamodb-local:8000"),
			},
		}))
		fmt.Println("Iniciando Sessão local")
		return sess
	}

}

func ListTable(session *session.Session) map[string]string {
	svc := dynamodb.New(session)

	// create the input configuration instance
	input := &dynamodb.ListTablesInput{}
	var tableNameMap = make(map[string]string)

	for {
		// Get the list of tables
		result, err := svc.ListTables(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case dynamodb.ErrCodeInternalServerError:
					fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return tableNameMap
		}

		for _, n := range result.TableNames {
			tableNameMap[*n] = *n
		}

		// assign the last read tablename as the start for our next call to the ListTables function
		// the maximum number of table names returned in a call is 100 (default), which requires us to make
		// multiple calls to the ListTables function to retrieve all table names
		input.ExclusiveStartTableName = result.LastEvaluatedTableName

		if result.LastEvaluatedTableName == nil {
			break
		}
	}
	return tableNameMap
}

func CreateTable() {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:   aws.String("us-east-1"),
			Endpoint: aws.String("http://localhost:8000"),
		},
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Create table Movies
	tableName := "Movies"

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Year"),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String("Title"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Year"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("Title"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	}

	_, err := svc.CreateTable(input)
	if err != nil {
		log.Fatalf("Got error calling CreateTable: %s", err)
	}

	fmt.Println("Created the table", tableName)

}

func CreateHttpTable(session *session.Session, tablename string) (bool, error) {

	// Create DynamoDB client
	svc := dynamodb.New(session)

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("httpCode"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("httpCode"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(tablename),
	}
	_, err := svc.CreateTable(input)
	if err != nil {
		return false, err
	}

	log.Println("[INFO] Tabela criada : ", tablename)
	return true, nil
}

func AddPutRequestSlice(sliceRequest *[]*dynamodb.WriteRequest, key int, value string) {
	formatedVal := dynamodb.WriteRequest{
		PutRequest: &dynamodb.PutRequest{
			Item: map[string]*dynamodb.AttributeValue{
				"httpCode": {
					N: aws.String(strconv.Itoa(key)),
				},
				"description": {
					S: aws.String(value),
				},
			},
		},
	}
	*sliceRequest = append(*sliceRequest, &formatedVal)
}

func BatchWriteItem(sess *session.Session, putRequestSlice *[]*dynamodb.WriteRequest, tablename string) {
	svc := dynamodb.New(sess)

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			tablename: *putRequestSlice,
		},
	}

	_, err := svc.BatchWriteItem(input)
	if err != nil {
		log.Println("Erro ao executar Insert Massivo ", err)
	}
	log.Printf("%v Itens inseridos com sucesso", len(*putRequestSlice))
}
