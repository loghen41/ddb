package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Update struct {
	api                                 dynamodbiface.DynamoDBAPI
	spec                                *tableSpec
	hashKey                             interface{}
	rangeKey                            interface{}
	consistentRead                      bool
	consumed                            *ConsumedCapacity
	err                                 error
	expr                                *expression
	returnValuesOnConditionCheckFailure string
}

func (u *Update) UpdateItemInput() (*dynamodb.UpdateItemInput, error) {
	key, err := makeKey(u.spec, u.hashKey, u.rangeKey)
	if err != nil {
		return nil, err
	}

	conditionExpression := u.expr.ConditionExpression()
	updateExpression := u.expr.UpdateExpression()
	return &dynamodb.UpdateItemInput{
		ConditionExpression:       conditionExpression,
		ExpressionAttributeNames:  u.expr.Names,
		ExpressionAttributeValues: u.expr.Values,
		Key:                       key,
		ReturnConsumedCapacity:    aws.String(dynamodb.ReturnConsumedCapacityTotal),
		TableName:                 aws.String(u.spec.TableName),
		UpdateExpression:          updateExpression,
	}, nil
}

func (u *Update) Tx() (*dynamodb.TransactWriteItem, error) {
	input, err := u.UpdateItemInput()
	if err != nil {
		return nil, err
	}

	writeItem := dynamodb.TransactWriteItem{
		Update: &dynamodb.Update{
			ConditionExpression:       input.ConditionExpression,
			ExpressionAttributeNames:  input.ExpressionAttributeNames,
			ExpressionAttributeValues: input.ExpressionAttributeValues,
			Key:                       input.Key,
			TableName:                 input.TableName,
			UpdateExpression:          input.UpdateExpression,
		},
	}
	if v := u.returnValuesOnConditionCheckFailure; v != "" {
		writeItem.Update.ReturnValuesOnConditionCheckFailure = aws.String(v)
	}

	return &writeItem, nil
}

func (u *Update) Add(expr string, values ...interface{}) *Update {
	if err := u.expr.Add(expr, values...); err != nil {
		u.err = err
	}

	return u
}

func (u *Update) Condition(expr string, values ...interface{}) *Update {
	if err := u.expr.Condition(expr, values...); err != nil {
		u.err = err
	}

	return u
}

func (u *Update) Delete(expr string, values ...interface{}) *Update {
	if err := u.expr.Delete(expr, values...); err != nil {
		u.err = err
	}

	return u
}

func (u *Update) Range(rangeKey interface{}) *Update {
	u.rangeKey = rangeKey
	return u
}

func (u *Update) Remove(expr string, values ...interface{}) *Update {
	if err := u.expr.Remove(expr, values...); err != nil {
		u.err = err
	}

	return u
}

func (u *Update) ReturnValuesOnConditionCheckFailure(value string) *Update {
	u.returnValuesOnConditionCheckFailure = value
	return u
}

func (u *Update) Set(expr string, values ...interface{}) *Update {
	if err := u.expr.Set(expr, values...); err != nil {
		u.err = err
	}

	return u
}

func (u *Update) RunWithContext(ctx context.Context) error {
	if u.err != nil {
		return u.err
	}

	input, err := u.UpdateItemInput()
	if err != nil {
		return err
	}

	output, err := u.api.UpdateItemWithContext(ctx, input)
	if err != nil {
		return err
	}

	u.consumed.add(output.ConsumedCapacity)
	return nil
}

func (u *Update) Run() error {
	return u.RunWithContext(defaultContext)
}

func (t *Table) Update(hashKey interface{}) *Update {
	return &Update{
		api:      t.ddb.api,
		spec:     t.spec,
		hashKey:  hashKey,
		consumed: t.consumed,
		expr: &expression{
			attributes: t.spec.Attributes,
		},
	}
}
