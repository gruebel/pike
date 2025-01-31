package pike

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

// Compare IAC codebase to AWS policy
func Compare(directory string, arn string) error {

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := iam.NewFromConfig(cfg)

	Version := GetVersion(client, arn)
	Policy, _ := GetPolicyVersion(client, arn, Version)

	iacPolicy, _ := MakePolicy(directory, "json")
	Sorted, _ := SortActions(iacPolicy)

	// iam versus iac
	fmt.Printf("IAM Policy %s versus Local %s \n", arn, directory)
	_, err = CompareIAMPolicy(Policy, string(Sorted))

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// CompareIAMPolicy takes to IAm policies and compares
func CompareIAMPolicy(Policy string, OldPolicy string) (bool, error) {

	differ := diff.New()
	d, err := differ.Compare([]byte(Policy), []byte(OldPolicy))

	if err != nil {
		return false, err
	}

	if d.Modified() {
		var aJSON map[string]interface{}
		err = json.Unmarshal([]byte(Policy), &aJSON)

		if err != nil {
			return false, err
		}

		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		formatter := formatter.NewAsciiFormatter(aJSON, config)
		diffString, err := formatter.Format(d)

		if err != nil {
			return false, err
		}

		fmt.Print(diffString)
		return true, nil
	}

	return false, nil
}
