package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"slowquery/pkg/awsapi"
	"slowquery/pkg/mysqllog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/slack-go/slack"
)

var (
	keySlackIncomingWebhookURL string
	slackIncomingWebhookURL    string
)

func init() {
	keySlackIncomingWebhookURL = os.Getenv("KEY_SLACK_INCOMING_WEBHOOK_URL")
	log.Printf("webhook確認: %v", keySlackIncomingWebhookURL)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(fmt.Sprintf("確認: %v", err))
	}
	ssmGetParameterAPI := ssm.NewFromConfig(cfg, func(o *ssm.Options) {
		o.Region = os.Getenv("AWS_REGION")
	})
	slackIncomingWebhookURL, err = awsapi.GetParameter(context.TODO(), ssmGetParameterAPI, &ssm.GetParameterInput{
		Name:           aws.String(keySlackIncomingWebhookURL),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Fatal(fmt.Errorf("failed awsapi.GetParameter: %v", err))
	}
}

// handler is our lambda handler invoked by the `lambda.Start` function call
func handler(ctx context.Context, event events.CloudwatchLogsEvent) error {
	// parse CloudWatch Logs Event
	data, _ := event.AWSLogs.Parse()

	logStream := data.LogStream

	log.Printf("logStreaの中を確認: %v", logStream)
	log.Printf("logEventの中身を確認: %v", data.LogEvents)
	for _, e := range data.LogEvents {
		message := e.Message

		reader := bytes.NewBufferString(message)
		q := mysqllog.Parser(reader)
		// Time もしくは Fingerprint でデータをマスクしたクエリが取得できない場合、以降処理を実施しない
		if q.Time == "" || q.Fingerprint == "" {
			continue
		}

		msg := slack.WebhookMessage{
			Attachments: []slack.Attachment{
				{
					Pretext: fmt.Sprintf("*Slow Query* on *%s*",
						logStream,
					),
					Color:      "#ba0000",
					AuthorName: "SlowQuery",
					AuthorIcon: "https://i.imgur.com/KzglgBi.png",
					Text:       fmt.Sprintf("```%s```", q.Fingerprint),
					Fields: []slack.AttachmentField{
						{
							Title: "Time",
							Value: q.Time,
							Short: true,
						},
						{
							Title: "User@Host",
							Value: fmt.Sprintf("%s@%s", q.User, q.Host),
							Short: true,
						},
						{
							Title: "Query_time",
							Value: q.QueryTime,
							Short: true,
						},
						{
							Title: "Lock_time",
							Value: q.LockTime,
							Short: true,
						},
						{
							Title: "Rows_sent",
							Value: q.RowsSent,
							Short: true,
						},
						{
							Title: "Rows_examined",
							Value: q.RowsExamined,
							Short: true,
						},
					},
				},
			},
		}
		if err := slack.PostWebhook(slackIncomingWebhookURL, &msg); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
