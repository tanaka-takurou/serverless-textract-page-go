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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

type APIResponse struct {
	Message  string `json:"message"`
}

type Response events.APIGatewayProxyResponse

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
		textractClient = getTextractClient(ctx)
	}

	input := &textract.AnalyzeDocumentInput{
		Document: &types.Document{
			Bytes: data,
		},
		FeatureTypes: []types.FeatureType{types.FeatureTypeTables},
	}
	res, err2 := textractClient.AnalyzeDocument(ctx, input)
	if err2 != nil {
		return "", err2
	}
	if len(res.Blocks) < 1 {
		return "No Document", nil
	}
	var wordList []string
	for _, v := range res.Blocks {
		if v.BlockType == types.BlockTypeWord || v.BlockType == types.BlockTypeLine {
			wordList = append(wordList, aws.ToString(v.Text))
		}
	}
	results, err3 := json.Marshal(wordList)
	if err3 != nil {
		return "", err3
	}
	return string(results), nil
}

func getTextractClient(ctx context.Context) *textract.Client {
	return textract.NewFromConfig(getConfig(ctx))
}

func getConfig(ctx context.Context) aws.Config {
	var err error
	newConfig, err := config.LoadDefaultConfig(ctx)
	newConfig.Region = "us-west-2"
	if err != nil {
		log.Print(err)
	}
	return newConfig
}

func main() {
	lambda.Start(HandleRequest)
}
