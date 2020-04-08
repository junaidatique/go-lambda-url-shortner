package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"	
	"crypto/sha256"
	"encoding/hex"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Link struct to hold info about new link that is returned from dynamoDB
type Link struct {
	ShortLink string
	Hash      string
	LongURL   string
}
// LinkAnalytics struct to hold info about link analytics
type LinkAnalytics struct {
	RequestID string
	ShortLink string
	SourceIP  string
	UserAgent string
}

// BodyRequest is our self-made struct to process JSON request from Client
type BodyRequest struct {
	RequestLongURL string `json:"LongURL"`
}

// BodyResponse is our self-made struct to build response for Client
type BodyResponse struct {
	ResponseShortLink string `json:"ShortLink"`
}

// CalcBase64 function returns the auto increment value into base 62 string
func CalcBase64() string {
	Client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL_HOST"),
		Password:     os.Getenv("REDIS_PASSWORD"),		
		DB:       0, 
	})
	var ShortLinkRedisVal int64
	
	Result, Err := Client.Get(os.Getenv("URL_COUNTER_KEY")).Result()
	if Err != nil {
		// leave this one as the value for key is not defined yet. 
		fmt.Println("Key value is not defined yet. ")
	}
	if Result == "" {
		Err := Client.Set(os.Getenv("URL_COUNTER_KEY"), os.Getenv("COUNTER_START_VALUE"), 0).Err()
		if Err != nil {
			panic(Err)
		}		 
		IntVal, Err := strconv.ParseInt(os.Getenv("COUNTER_START_VALUE"), 10, 64)
		ShortLinkRedisVal = IntVal
	} else {
		Result, Err := Client.Incr(os.Getenv("URL_COUNTER_KEY")).Result()
		if Err != nil {
				panic(Err)
		}	
		ShortLinkRedisVal = Result
	}	
	fmt.Println(ShortLinkRedisVal)
	Count := 0
	const Base62 string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	BaseArray := strings.SplitAfter(Base62, "")	
	Dividend := ShortLinkRedisVal
	ShortLinkResponse := make([]string, 10)	
	var Remainder int64
	for (Dividend > 0) {
		Remainder = Dividend % 62		
		Dividend = (Dividend / 62)		
		ShortLinkResponse[Count] = BaseArray[Remainder]
		Count = Count + 1
	}	
	Response := strings.Join(ShortLinkResponse, "")
	return Response
}
// GetItemFromDynamoDB query the DynamoDB 
func GetItemFromDynamoDB(GetItemKey string, GetItemVal string) (Link, error) {
	// Initialize a session that the SDK will use to load	
	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))
	// Create DynamoDB client
	svc := dynamodb.New(session)
	tableName := "Link"

	Result, Err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			GetItemKey: {
				S: aws.String(GetItemVal),
			},
		},
	})
	if Err != nil {
		fmt.Println(Err.Error())
		return Link{}, Err
	}
	item := Link{}
	Err = dynamodbattribute.UnmarshalMap(Result.Item, &item)		
	if Err != nil {
		return Link{}, Err
	}	
	return item, nil
}
// GetShortLinkFromHash get hash from link table
func GetShortLinkFromHash(Hash string) (Link, error) {
	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))
	// Create DynamoDB client
	svc := dynamodb.New(session)
	tableName := "Link"
	Result, Err := svc.Query(&dynamodb.QueryInput{
		TableName: aws.String(tableName),
		IndexName: aws.String("HashIndex"),
    KeyConditions: map[string]*dynamodb.Condition{
			"Hash": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(Hash),
					},
				},
			},
    },
	})
	if Err != nil {
		return Link{}, Err
	}	
	
	items := []Link{}
	Err = dynamodbattribute.UnmarshalListOfMaps(Result.Items, &items)
	if Err != nil {
		return Link{}, Err
	}
	if (len(items) > 0) {
		return items[0], nil
	} else {
		return Link{}, nil
	}
}

// PutLinkAnalyticsInDynamoDB query the DynamoDB 
func PutLinkAnalyticsInDynamoDB(request events.APIGatewayProxyRequest, ShortLink string) (LinkAnalytics, error) {

	item := LinkAnalytics{
		RequestID: request.RequestContext.RequestID,
		ShortLink: ShortLink,
		SourceIP: request.RequestContext.Identity.SourceIP,
		UserAgent: request.RequestContext.Identity.UserAgent,
	}
	

	// Initialize a session that the SDK will use to load	
	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))
	// Create DynamoDB client
	svc := dynamodb.New(session)
	tableName := "LinkAnalytics"
	PutItemVal, Err := dynamodbattribute.MarshalMap(item)
	input := &dynamodb.PutItemInput{
    Item:      PutItemVal,
    TableName: aws.String(tableName),
	}
	_, Err = svc.PutItem(input)
	if Err != nil {
		fmt.Println(Err.Error())
		return LinkAnalytics{}, Err
	}
	return item, nil
}
// PutLinkItemFromDynamoDB query the DynamoDB 
func PutLinkItemFromDynamoDB(PutItemStruct Link) (Link, error) {
	// Initialize a session that the SDK will use to load	
	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))
	// Create DynamoDB client
	svc := dynamodb.New(session)
	tableName := "Link"
	PutItemVal, Err := dynamodbattribute.MarshalMap(PutItemStruct)
	input := &dynamodb.PutItemInput{
    Item:      PutItemVal,
    TableName: aws.String(tableName),
	}
	_, Err = svc.PutItem(input)
	if Err != nil {
		fmt.Println(Err.Error())
		return Link{}, Err
	}
	return PutItemStruct, nil
}



// Handler is your Lambda function handler
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {		
	

	if request.HTTPMethod == "GET" {
		fmt.Printf("GET METHOD\n")
		ShortLink := request.QueryStringParameters["ShortLink"]
		PutLinkAnalyticsInDynamoDB(request, ShortLink)
		fmt.Printf("short link %s ", ShortLink)
		if ShortLink == "" {
			return events.APIGatewayProxyResponse{
				Body: "{\"error\" : \"Short link not provided.\"} ", 
				StatusCode: 400,
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
				},
			}, nil
		}
		item, Err := GetItemFromDynamoDB("ShortLink", ShortLink)
		if Err != nil {
			ErrorMessage := fmt.Sprintf(" { \"error\" : \"%s\" } ", Err.Error())
			return events.APIGatewayProxyResponse{
				Body: ErrorMessage, 
				StatusCode: 400,
				Headers: map[string]string{
					"Access-Control-Allow-Origin" : "*",
				},
			}, nil
		}
		if (item == Link{}) {
			ErrorMessage := fmt.Sprintf(" { \"error\" : \"Short link not found.\" } ")
			return events.APIGatewayProxyResponse{
				Body: ErrorMessage, 
				StatusCode: 400,
				Headers: map[string]string{
					"Access-Control-Allow-Origin" : "*",
				},
			}, nil
		}
		message := fmt.Sprintf(" { \"ShortLink\" : \"%s\", \"LongURL\" : \"%s\" } ", item.ShortLink, item.LongURL)
		return events.APIGatewayProxyResponse{
			Body: message, 
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin" : "*",
			},
		}, nil
	} else if request.HTTPMethod == "POST" {
		fmt.Printf("POST METHOD\n")

		// BodyRequest will be used to take the json response from client and build it
		bodyRequest := BodyRequest{
			RequestLongURL: "",
		}
		fmt.Printf("bodyRequest: %+v\n", bodyRequest)
		fmt.Printf("request.Body: %+v\n", request.Body)
		// Unmarshal the json, return 404 if Error
		Err := json.Unmarshal([]byte(request.Body), &bodyRequest)
		if Err != nil {
			return events.APIGatewayProxyResponse{
				Body: Err.Error(), 
				StatusCode: 404,
				Headers: map[string]string{
					"Access-Control-Allow-Origin" : "*",
				},
			}, nil
		}
		fmt.Printf("bodyRequest: %+v\n", bodyRequest)
		LongURLErr := validation.Validate(bodyRequest.RequestLongURL,
						validation.Required,       // not empty						
						is.URL,                    // is a valid URL
					)
		fmt.Printf("LongURLErr: %s\n", LongURLErr)
		if LongURLErr != nil {
			return events.APIGatewayProxyResponse{
				Body: "{\"error\" : \"URL is not valid\"}", 
				StatusCode: 404,
				Headers: map[string]string{
					"Access-Control-Allow-Origin" : "*",
				},
			}, nil
		}
		fmt.Printf("RequestLongURL: %s\n", bodyRequest.RequestLongURL)		
		LongURLHashBytes := sha256.Sum256([]byte(bodyRequest.RequestLongURL))
		LongURLHash := hex.EncodeToString(LongURLHashBytes[:])
		fmt.Printf("LongURLHash %s\n", LongURLHash)
		item, Err := GetShortLinkFromHash(LongURLHash)
		if Err != nil {
			ErrorMessage := fmt.Sprintf(" { \"error\" : \"%s\" } ", Err.Error())
			return events.APIGatewayProxyResponse{
				Body: ErrorMessage, 
				StatusCode: 400,
				Headers: map[string]string{
					"Access-Control-Allow-Origin" : "*",
				},
			}, nil
		}
		fmt.Printf("item %s\n", item)
		bodyResponse := BodyResponse{}
		if (item == Link{}) {
			ShortLinkRedisVal := CalcBase64()	
			fmt.Println(ShortLinkRedisVal)
			item := Link{
				ShortLink: ShortLinkRedisVal,
				Hash: LongURLHash,
				LongURL: bodyRequest.RequestLongURL,
			}
			_, Err = PutLinkItemFromDynamoDB(item)
			if Err != nil {
				ErrorMessage := fmt.Sprintf(" { \"error\" : \"%s\" } ", Err.Error())
				return events.APIGatewayProxyResponse{
					Body: ErrorMessage, 
					StatusCode: 400,
					Headers: map[string]string{
						"Access-Control-Allow-Origin" : "*",
					},
				}, nil
			}
			bodyResponse = BodyResponse{
				ResponseShortLink: ShortLinkRedisVal,
			}
		} else {
			// We will build the BodyResponse and send it back in json form
			bodyResponse = BodyResponse{
				ResponseShortLink: item.ShortLink,
			}
			
		}
		response, Err := json.Marshal(&bodyResponse)
		if Err != nil {
			ErrorMessage := fmt.Sprintf(" { \"error\" : \"%s\" } ", Err.Error())
			return events.APIGatewayProxyResponse{
				Body: ErrorMessage, 
				StatusCode: 404,
				Headers: map[string]string{
					"Access-Control-Allow-Origin" : "*",
				},
			}, nil
		}
		
		return events.APIGatewayProxyResponse{
			Body: string(response), 
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin" : "*",
			},
		}, nil
	} else {
		fmt.Printf("NEITHER\n")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin" : "*",
			},
		}, nil
	}
}

func main() {
	log.Printf("Start lambda")
	lambda.Start(Handler)
}
