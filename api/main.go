package main

import (
	"fmt"
	"log"
	"strings"
	"context"
	"encoding/json"
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type APIResponse struct {
	Message  string `json:"message"`
}

type Response events.APIGatewayProxyResponse

const layout string = "2006-01-02 15:04"

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	d := make(map[string]string)
	json.Unmarshal([]byte(request.Body), &d)
	if v, ok := d["action"]; ok {
		switch v {
		case "analyzedocument" :
			if i, ok := d["image"]; ok {
				r, e := analyzeDocument(i)
				if e != nil {
					err = e
				} else {
					jsonBytes, _ = json.Marshal(APIResponse{Message: r})
				}
			}
		}
	}
	log.Print(request.RequestContext.Identity.SourceIP)
	if err != nil {
		log.Print(err)
		jsonBytes, _ = json.Marshal(APIResponse{Message: fmt.Sprint(err)})
		return Response{
			StatusCode: 500,
			Body: string(jsonBytes),
		}, nil
	}
	return Response {
		StatusCode: 200,
		Body: string(jsonBytes),
	}, nil
}

func analyzeDocument(img string)(string, error) {
	b64data := img[strings.IndexByte(img, ',')+1:]
	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		log.Print(err)
		return "", err
	}
	svc := textract.New(session.New(), &aws.Config{
		Region: aws.String("us-west-2"),
	})

	input := &textract.AnalyzeDocumentInput{
		Document: &textract.Document{
			Bytes: data,
		},
		FeatureTypes: []*string{aws.String("TABLES")},
	}
	res, err2 := svc.AnalyzeDocument(input)
	if err2 != nil {
		log.Print(err2)
		return "", err2
	}
	if len(res.Blocks) < 1 {
		return "No Document", nil
	}
	var wordList []string
	for _, v := range res.Blocks {
		if aws.StringValue(v.BlockType) == "WORD" || aws.StringValue(v.BlockType) == "LINE" {
			wordList = append(wordList, aws.StringValue(v.Text))
		}
	}
	results, err3 := json.Marshal(wordList)
	if err3 != nil {
		log.Print(err3)
		return "", err3
	}
	return string(results), nil
}

func main() {
	lambda.Start(HandleRequest)
}
