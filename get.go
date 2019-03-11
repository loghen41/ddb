package ddb

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Get struct {
	api            dynamodbiface.DynamoDBAPI
	spec           *tableSpec
	hashKey        Value
	rangeKey       Value
	consistentRead bool
	consumed       *ConsumedCapacity
}

func (g *Get) makeGetItemInput() *dynamodb.GetItemInput {
	key := makeKey(g.spec, g.hashKey, g.rangeKey)
	return &dynamodb.GetItemInput{
		ConsistentRead:         aws.Bool(g.consistentRead),
		Key:                    key,
		TableName:              aws.String(g.spec.TableName),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}
}

func (g *Get) Range(value Value) *Get {
	g.rangeKey = value
	return g
}

func (g *Get) ConsistentRead(enabled bool) *Get {
	g.consistentRead = true
	return g
}

func (g *Get) ScanWithContext(ctx context.Context, v interface{}) error {
	input := g.makeGetItemInput()
	output, err := g.api.GetItemWithContext(ctx, input)
	if err != nil {
		return err
	}

	g.consumed.add(output.ConsumedCapacity)
	if len(output.Item) == 0 {
		return errorf(ErrItemNotFound, "item not found")
	}

	if err := dynamodbattribute.UnmarshalMap(output.Item, v); err != nil {
		return err
	}

	return nil
}

func (g *Get) Scan(v interface{}) error {
	return g.ScanWithContext(defaultContext, v)
}

func (t *Table) Get(hashKey Value) *Get {
	return &Get{
		api:      t.ddb.api,
		spec:     t.spec,
		hashKey:  hashKey,
		consumed: t.consumed,
	}
}

type Value struct {
	item *dynamodb.AttributeValue
}

func String(v string) Value {
	return Value{
		item: &dynamodb.AttributeValue{S: aws.String(v)},
	}
}

func Int64(v int64) Value {
	return Value{
		item: &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(v, 10))},
	}
}

func Raw(item *dynamodb.AttributeValue) Value {
	return Value{
		item: item,
	}
}
