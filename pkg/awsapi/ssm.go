package awsapi

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type SSMGetParameterAPI interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

func GetParameter(ctx context.Context, api SSMGetParameterAPI, input *ssm.GetParameterInput) (val string, err error) {
	r, err := api.GetParameter(ctx, input)
	if err != nil {
		return "", err
	}
	return *r.Parameter.Value, nil
}
