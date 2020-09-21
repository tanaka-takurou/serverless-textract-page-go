package main

import (
	"fmt"
	"log"
	"strings"
	"context"
	"encoding/json"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/textract"
)

type APIResponse struct {
	Message  string `json:"message"`
}

type Response events.APIGatewayProxyResponse

var cfg aws.Config
var textractClient *textract.Client

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	d := make(map[string]string)
	json.Unmarshal([]byte(request.Body), &d)
	if v, ok := d["action"]; ok {
		switch v {
		case "analyzedocument" :
			if i, ok := d["image"]; ok {
				r, e := analyzeDocument(ctx, i)
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

func analyzeDocument(ctx context.Context, img string)(string, error) {
	b64data := img[strings.IndexByte(img, ',')+1:]
	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		log.Print(err)
		return "", err
	}
	if textractClient == nil {
		cfg.Region = "us-west-2"
		textractClient = textract.New(cfg)
	}

	input := &textract.AnalyzeDocumentInput{
		Document: &textract.Document{
			Bytes: data,
		},
		FeatureTypes: []textract.FeatureType{textract.FeatureTypeTables},
	}
	req := textractClient.AnalyzeDocumentRequest(input)
	res, err2 := req.Send(ctx)
	if err2 != nil {
		return "", err2
	}
	if len(res.AnalyzeDocumentOutput.Blocks) < 1 {
		return "No Document", nil
	}
	var wordList []string
	for _, v := range res.AnalyzeDocumentOutput.Blocks {
		if v.BlockType == textract.BlockTypeWord || v.BlockType == textract.BlockTypeLine {
			wordList = append(wordList, aws.StringValue(v.Text))
		}
	}
	results, err3 := json.Marshal(wordList)
	if err3 != nil {
		return "", err3
	}
	return string(results), nil
}

func init() {
	var err error
	cfg, err = external.LoadDefaultAWSConfig()
	if err != nil {
		log.Print(err)
	}
}

func main() {
	lambda.Start(HandleRequest)
}
