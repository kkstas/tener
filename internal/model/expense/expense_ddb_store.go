package expense

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/kkstas/tjener/internal/helpers"
)

type DDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func getKey(vaultID, sk string) map[string]types.AttributeValue {
	PK, err := attributevalue.Marshal(buildPK(vaultID))
	if err != nil {
		panic(err)
	}
	SK, err := attributevalue.Marshal(sk)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"PK": PK, "SK": SK}
}

func NewDDBStore(tableName string, client *dynamodb.Client) *DDBStore {
	return &DDBStore{
		tableName: tableName,
		client:    client,
	}
}

func (es *DDBStore) marshal(
	pk string,
	sk string,
	name string,
	date string,
	amount float64,
	paymentMethod string,
	category string,
	createdAt string,
	userID string,
) (Expense, map[string]types.AttributeValue, error) {
	newExpense := Expense{
		PK:            pk,
		SK:            sk,
		Name:          name,
		Date:          date,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Category:      category,
		CreatedAt:     createdAt,
		CreatedBy:     userID,
	}
	item, err := attributevalue.MarshalMap(newExpense)
	return newExpense, item, err
}

func (es *DDBStore) Create(ctx context.Context, expenseFC Expense, userID, vaultID string) (Expense, error) {
	newExpense, item, err := es.marshal(
		buildPK(vaultID),
		expenseFC.SK,
		expenseFC.Name,
		expenseFC.Date,
		expenseFC.Amount,
		expenseFC.PaymentMethod,
		expenseFC.Category,
		expenseFC.CreatedAt,
		userID,
	)

	if err != nil {
		return Expense{}, fmt.Errorf("failed to marshal expense: %w", err)
	}

	_, err = es.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &es.tableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(SK)"),
	})

	if err != nil {
		return Expense{}, fmt.Errorf("failed to put item into DynamoDB: %w", err)
	}

	err = es.updateMonthlySum(ctx, vaultID, newExpense.Date, newExpense.Category)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to update monthly sum: %w", err)
	}

	return newExpense, nil
}

func (es *DDBStore) FindOne(ctx context.Context, sk, vaultID string) (Expense, error) {
	expense := Expense{PK: buildPK(vaultID), SK: sk}
	response, err := es.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &es.tableName,
		Key:       getKey(vaultID, sk),
	})

	if err != nil {
		return Expense{}, fmt.Errorf("GetItem DynamoDB operation failed for SK='%s': %w", sk, err)
	}

	if len(response.Item) == 0 {
		return Expense{}, &NotFoundError{SK: sk}
	}

	err = attributevalue.UnmarshalMap(response.Item, &expense)
	if err != nil {
		return Expense{}, fmt.Errorf("failed to unmarshal expense: %w", err)
	}

	return expense, nil
}

func (es *DDBStore) Update(ctx context.Context, expenseFU Expense, vaultID string) error {
	foundExpense, err := es.FindOne(ctx, expenseFU.SK, vaultID)
	if err != nil {
		return fmt.Errorf("failed to find expense for update: %w", err)
	}

	expenseFU.CreatedAt = foundExpense.CreatedAt
	expenseFU.CreatedBy = foundExpense.CreatedBy

	if expenseFU.SK == buildSK(expenseFU.Date, foundExpense.CreatedAt) {
		err = es.updateWithoutNewSK(ctx, expenseFU, vaultID)
	} else {
		err = es.updateWithNewSK(ctx, expenseFU, vaultID)
	}
	if err != nil {
		return fmt.Errorf("failed to update expense: %w", err)
	}

	err = es.updateMonthlySum(ctx, vaultID, expenseFU.Date, expenseFU.Category)
	if err != nil {
		return fmt.Errorf("failed to update monthly sum: %w", err)
	}

	return err
}

func (es *DDBStore) updateWithNewSK(ctx context.Context, expenseFU Expense, vaultID string) error {
	deleteItem := types.TransactWriteItem{
		Delete: &types.Delete{
			TableName:           aws.String(es.tableName),
			Key:                 getKey(vaultID, expenseFU.SK),
			ConditionExpression: aws.String("attribute_exists(SK)"),
		},
	}

	expense, item, err := es.marshal(
		buildPK(vaultID),
		buildSK(expenseFU.Date, expenseFU.CreatedAt),
		expenseFU.Name,
		expenseFU.Date,
		expenseFU.Amount,
		expenseFU.PaymentMethod,
		expenseFU.Category,
		expenseFU.CreatedAt,
		expenseFU.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to marshal expense: %w", err)
	}

	putItem := types.TransactWriteItem{
		Put: &types.Put{
			TableName: aws.String(es.tableName),
			Item:      item,
		},
	}

	_, err = es.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{deleteItem, putItem},
	})

	if err != nil {
		var transactionErr *types.TransactionCanceledException

		if errors.As(err, &transactionErr) {
			for _, reason := range transactionErr.CancellationReasons {
				if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
					return &NotFoundError{SK: expense.SK}
				}
			}
		}
		return fmt.Errorf("failed to update expense atomically: %w", err)
	}

	return nil
}

func (es *DDBStore) updateWithoutNewSK(ctx context.Context, expenseFU Expense, vaultID string) error {
	update := expression.
		Set(expression.Name("name"), expression.Value(expenseFU.Name)).
		Set(expression.Name("category"), expression.Value(expenseFU.Category)).
		Set(expression.Name("amount"), expression.Value(expenseFU.Amount)).
		Set(expression.Name("paymentMethod"), expression.Value(expenseFU.PaymentMethod))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return fmt.Errorf("failed to build expression for update: %w", err)
	}

	_, err = es.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &es.tableName,
		Key:                       getKey(vaultID, expenseFU.SK),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
		ConditionExpression:       aws.String("attribute_exists(SK)"),
	})

	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return &NotFoundError{SK: expenseFU.SK}
		}
		return fmt.Errorf("failed to update expense: %w", err)
	}

	return nil
}

func (es *DDBStore) Delete(ctx context.Context, sk, vaultID string) error {
	exp, err := es.FindOne(ctx, sk, vaultID)
	if err != nil {
		return &NotFoundError{SK: sk}
	}

	_, err = es.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &es.tableName,
		Key:       getKey(vaultID, sk),
	})

	if err != nil {
		return fmt.Errorf("failed to delete expense with SK=%q from table: %w", sk, err)
	}

	err = es.updateMonthlySum(ctx, vaultID, exp.Date, exp.Category)
	if err != nil {
		return fmt.Errorf("failed to update monthly sum: %w", err)
	}

	return err
}

func (es *DDBStore) updateMonthlySum(ctx context.Context, vaultID, date, category string) error {
	yearAndMonth := date[:7]
	if !helpers.IsValidYYYYMM(yearAndMonth) {
		return fmt.Errorf("error: expected date in format YYYY-MM, got %s", date)
	}

	sum, err := es.calcMonthlySum(ctx, vaultID, category, date)
	if err != nil {
		return fmt.Errorf("failed to calculate monthly sum: %w", err)
	}

	item, err := attributevalue.MarshalMap(
		MonthlySum{
			PK:  buildMonthlySumPK(vaultID),
			SK:  buildMonthlySumSK(yearAndMonth, category),
			Sum: sum,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal monthly sum: %w", err)
	}

	_, err = es.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &es.tableName,
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to put monthly sum into DynamoDB: %w", err)
	}

	return nil
}

func (es *DDBStore) calcMonthlySum(ctx context.Context, activeVault, category, dateStr string) (float64, error) {
	from, to := helpers.GetFirstAndLastDayOfMonth(dateStr)
	thisMonthCategoryExpenses, err := es.Query(ctx, from, to, []string{category}, activeVault)
	if err != nil {
		return 0, err
	}

	var sum float64
	for _, val := range thisMonthCategoryExpenses {
		sum += val.Amount
	}

	return math.Floor(sum*100) / 100, nil
}

func (es *DDBStore) GetMonthlySums(ctx context.Context, monthsAgo int, vaultID string) ([]MonthlySum, error) {
	from := helpers.MonthsAgo(monthsAgo)[:7]

	keyCond := expression.
		Key("PK").Equal(expression.Value(buildMonthlySumPK(vaultID))).
		And(expression.Key("SK").GreaterThanEqual(expression.Value(from)))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression for monthlysums query %w", err)
	}

	var monthlySums []MonthlySum

	queryInput := dynamodb.QueryInput{
		TableName:                 &es.tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	queryPaginator := dynamodb.NewQueryPaginator(es.client, &queryInput)

	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query for expenses: %w", err)
		}

		var resMonthlySums []MonthlySum
		err = attributevalue.UnmarshalListOfMaps(response.Items, &resMonthlySums)

		if err != nil {
			return monthlySums, fmt.Errorf("failed to unmarshal query response %w", err)
		}

		monthlySums = append(monthlySums, resMonthlySums...)
	}

	return monthlySums, nil
}

// Retrieves expenses between the given `from` and `to` YYYY-MM-DD dates (inclusive).
func (es *DDBStore) Query(ctx context.Context, from, to string, categories []string, vaultID string) ([]Expense, error) {
	daysDiff, err := helpers.DaysBetween(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get number of days between 'from' and 'to' date: %w", err)
	}
	if daysDiff < minQueryRangeDaysDiff || daysDiff > maxQueryRangeDaysDiff {
		return nil, fmt.Errorf(
			"invalid difference between 'from' and 'to' date; got=%d, max=%d, min=%d",
			daysDiff,
			minQueryRangeDaysDiff,
			maxQueryRangeDaysDiff,
		)
	}

	dayAfterTo, err := helpers.NextDay(to)
	if err != nil {
		return nil, fmt.Errorf("failed to get next day for date '%s': %w", to, err)
	}

	keyCond := expression.
		Key("PK").Equal(expression.Value(buildPK(vaultID))).
		And(expression.Key("SK").Between(expression.Value(from), expression.Value(dayAfterTo)))

	exprBuilder := expression.NewBuilder().WithKeyCondition(keyCond)

	if len(categories) > 0 {
		categoryCondition := expression.Name("category").In(expression.Value(categories[0]))
		for _, category := range categories[1:] {
			categoryCondition = categoryCondition.Or(expression.Name("category").In(expression.Value(category)))
		}
		exprBuilder = exprBuilder.WithFilter(categoryCondition)
	}

	expr, err := exprBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression for query %w", err)
	}

	return es.query(ctx, expr)
}

func (es *DDBStore) query(ctx context.Context, expr expression.Expression) ([]Expense, error) {
	var expenses []Expense

	queryInput := dynamodb.QueryInput{
		TableName:                 &es.tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	queryPaginator := dynamodb.NewQueryPaginator(es.client, &queryInput)

	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query for expenses: %w", err)
		}

		var resExpenses []Expense
		err = attributevalue.UnmarshalListOfMaps(response.Items, &resExpenses)

		if err != nil {
			return expenses, fmt.Errorf("failed to unmarshal query response %w", err)
		}

		expenses = append(expenses, resExpenses...)
	}

	return expenses, nil
}
