package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

type CommentModel struct {
	SVC ItemService
}

type Comment struct {
	PK      string `dynamodbav:"PK"`
	SK      string `dynamodbav:"SK"`
	GSI4PK  string `dynamodbav:"GSI4PK"`
	GSI4SK  string `dynamodbav:"GSI4SK"`
	Content string `dynamodbav:"content"`
}

// PUT - Comment on a molt by crab
func (m CommentModel) Insert(c *Comment, molt *Molt, crabUsername string) error {
	comment, err := attributevalue.MarshalMap(c)
	if err != nil {
		fmt.Println("ERR marshalling: ", err)
		panic(err)
	}
	ownerID := molt.PK[2:]
	fmt.Println("username is commenting on...", crabUsername)
	notification, err := attributevalue.MarshalMap(
		&Notification{
			PK:       fmt.Sprintf("N#%s", ownerID),                      // alert original author of molt
			SK:       fmt.Sprintf("N#%s#%s#%s", ownerID, "MC", molt.ID), // what if multiple comments then it would overwrite?
			UserName: crabUsername,
			Content:  c.Content,
			Viewed:   false,
			TTL:      fmt.Sprintf("%d", time.Now().Add(time.Hour*24*7).Unix()), // delete notifs in a week to keep table smaller
		})
	if err != nil {
		fmt.Println("Notification ERR: ", err)
		panic(err)
	}
	tItems := make([]types.TransactWriteItem, 0)
	tw1 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                comment,
			TableName:           aws.String(TableName),
			ConditionExpression: aws.String("attribute_not_exists(PK)"),
		},
	}
	tw2 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: molt.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: molt.SK,
				},
			},
			ConditionExpression: aws.String("attribute_exists(PK)"),
			TableName:           aws.String(TableName),
			UpdateExpression:    aws.String("set #comment_count = #comment_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#comment_count": "comment_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tw3 := types.TransactWriteItem{
		Put: &types.Put{
			Item:      notification,
			TableName: aws.String(TableName),
		},
	}

	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)
	tItems = append(tItems, tw3)

	_, err = m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
	}
	return err

}

// GET - Shows crabs comments in general
func (m CommentModel) Show(crabID string) ([]Comment, error) {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(5),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: fmt.Sprintf("MC#%s", crabID)},
		},
		ScanIndexForward: aws.Bool(false),
	})
	// update this for pagination
	comments := make([]Comment, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var comment []Comment
		err = attributevalue.UnmarshalListOfMaps(out.Items, &comment)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		comments = append(comments, comment...)

	}
	return comments, nil
}

// GET - Comments on a specific molt TODO add this for Remolts and Likes
func (m CommentModel) On(moltID string) ([]Comment, error) {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("GSI4"),
		Limit:                  aws.Int32(5),
		KeyConditionExpression: aws.String("GSI4PK = :gsi4pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi4pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("MC#%s", moltID)},
		},
	})
	// update this for pagination
	comments := make([]Comment, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var comment []Comment
		err = attributevalue.UnmarshalListOfMaps(out.Items, &comment)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		comments = append(comments, comment...)

	}
	return comments, nil
}
