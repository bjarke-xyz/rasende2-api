package ai

import (
	"context"
	"fmt"
	"log"

	"github.com/bjarke-xyz/rasende2-api/pkg"
	"github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	appContext *pkg.AppContext
	client     *openai.Client
}

func NewOpenAIClient(appContext *pkg.AppContext) *OpenAIClient {
	client := openai.NewClient(appContext.Config.OpenAIAPIKey)
	return &OpenAIClient{
		appContext: appContext,
		client:     client,
	}
}

func (o *OpenAIClient) GenerateArticleTitles(ctx context.Context, siteName string, siteDescription string, previousTitles []string, newTitlesCount int, temperature float32) (*openai.ChatCompletionStream, error) {
	if newTitlesCount > 10 {
		newTitlesCount = 10
	}
	previousTitlesStr := ""
	model := "gpt-4o-mini"
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}
	var token []int
	previousTitlesCount := 0
	for _, prevTitle := range previousTitles {
		previousTitlesCount++
		tmpStr := previousTitlesStr + "\n" + prevTitle
		token = tkm.Encode(tmpStr, nil, nil)
		if len(token) > 3000 {
			break
		}
		previousTitlesStr = tmpStr
	}
	log.Printf("GenerateArticleTitles - site: %v, tokens: %v, previousTitles: %v", siteName, len(token), previousTitlesCount)
	req := openai.ChatCompletionRequest{
		Model:       model,
		Temperature: temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("Du er en journalist på mediet %v. %v. \nDu vil få stillet en række tidligere overskrifter til rådighed. Find på %v nye overskrifter, der minder om de overskrifter du får. Begynd hver overskrift på en ny linje. Start hver linje med et mellemrum (' '). Returner kun overskrifter, intet andet. Lav højest %v overskrifter.", siteName, siteDescription, newTitlesCount, newTitlesCount),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: previousTitlesStr,
			},
		},
		Stream: true,
	}
	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}
	return stream, err
}

func (o *OpenAIClient) GenerateArticleContent(ctx context.Context, siteName string, siteDescription string, articleTitle string, temperature float32) (*openai.ChatCompletionStream, error) {
	log.Printf("GenerateArticleContent - site: %v, title: %v, temperature: %v", siteName, articleTitle, temperature)
	model := openai.GPT3Dot5Turbo
	req := openai.ChatCompletionRequest{
		Model:       model,
		Temperature: temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("Du er en journalist på mediet %v. %v. \nDu vil få en overskrift, og du skal skrive en artikel der passer til den overskrift.", siteName, siteDescription),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: articleTitle,
			},
		},
		Stream: true,
	}
	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}
	return stream, err
}
