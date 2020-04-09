# URL Shortener in Golang with AWS Lambda and DynamoDB

This application is built in Golang with terraform to setup AWS Resources. 

## Installation

In order to install the app please clone the project. change the file `infra/sample.secret.tfvars` to `secret.tfvars`.



## Commands

```
make build
```
This command will generate application binary and store in `build/bin/app`.
```
make init
```
This command will execute the `terraform init` command on a folder name `infra`, this command requires to run once to initialize everything terraform require to provision your infrastructure.
```
make plan:
```
This command will run the `terraform plan` command on a folder name `infra`, this command gives the detail about what will happen in your infrastructure.
```
make apply
```
This command will run the `terraform apply` the command to make changes to infrastructure. This command requires variables and it also auto-approve without human intervention. 
```
make destroy
```
This command will run the `terraform destroy` command on the folder name `infra`. This is the destroy command that will remove all the infrastructure component that was provision previously.


## Logic

In order to create a unique short url for there are many options. Whatever the option, we do have to query the DB to restrict the uniqueness of the short URL. Plus we also need to check that if the URL is existing in the DB,
we have to make another query. 

So if for example we are getting 1000rps, then foreach request we have to make at least 2 DB calls and this is the best case scenario that the short link is not present in the DB.

In order to reduce these DB calls, I purpose to use an algorithm that always generate a unique short link so that we don't have to query the DB. 

## Algorithm

We have 62 alpha numeric chars i.e. [a-z 0–9 A-Z]. we need to convert unique base10 value to base62 in order to get unique value for the URL. In order to get unique base10 value, I use Redis auto-increment feature (as it is atomic in nature). Using redis INCR function, I can get a seed number and generate a base62 string. Although this redis service is single point failure but we can use redis cluster for that. 

For this project I have initialized auto-increment variable which will generate 3 letter short link. we can adjust that according to the requirement. 

In order to check duplication of the URL and for better performance, we can encrypt the given URL through some encryption method i.e. md5, sha1 etc. For this project I am encrypting the given URL using sha256 and save the hash into DB. whenever user demand short link for the URL, we convert this URL into SHA256 hash and query the DB.

Now there is a tradeoff between two approaches. We can check for the duplication of URL in DB or we can allow duplication in order to save the cost of query for each DB operation. In order to create fully scalable URL shortener, we will not check for the duplication and as we always have unique short link we will save the DB operation cost and just save the record in the DB. But that totally depends on some the requirement such as number of calls per second, cost of storage and Db operations. 

## API

The API is developed in Golang using the AWS API Gateway packages. the API is single end point.

### GET Method
The get method returns the short link based on the input. It also saves the analytics information such as source IP, user browser etc. 

#### Payload
When no short link is provided.
```
{} or {"ShortLink": ""}
```
#### Response
```
{
  Body: {"error" : "Short link not provided."}, 
  StatusCode: 400,  
}
```

#### Payload
When short link does not exist
```
{"ShortLink": "invalid short link"}
```
#### Response
```
{
  Body: {"error" : "Short link does not exist."}, 
  StatusCode: 404,  
}
```
#### Payload
When valid short link is provided.
```
{"ShortLink": "qbA"}
```
#### Response
```
{
  Body: { "ShortLink" : "qbA", "LongURL" : "http:\\localhost" }, 
  StatusCode: 200,  
}
```
### POST Method
This method is used to generate new short link for the given URL. 

#### Payload
When valid short link is provided.
```
{} or {"LongURL":""} or {"LongURL": "invalidURL"}
```
#### Response
```
{
  Body: {"error" : "URL is not valid"}, 
  StatusCode: 400,  
}
```
#### Payload
When valid short link is provided.
```
{"LongURL": "http:\\localhost"}
```
#### Response
```
{
  Body: { "ShortLink" : "qbA" }, 
  StatusCode: 200,  
}
```

### DB
I have used dynamo DB for this project. For more information about DB structure please review `infra/dynamodb.tf` and `infra/dynammodb_link_analytics.tf` files. These files define 2 dynamo DB tables. One is named `Link` and it is used to store link data. The second table is named `LinkAnalytics` and it stores analytics about user clicks. 

### Terraform
This project uses Terraform to manage AWS infrastructure. 

### Known Issue
The API calls fails in React JS app when used with `axios`. I have tried to setup options method in terraform but its not working. 