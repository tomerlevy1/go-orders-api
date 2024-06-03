package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/tomerlevy1/go-orders-api/model"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("error marshalling order: %v", err)
	}

	key := orderIDKey(order.OrderID)

	txn := r.Client.TxPipeline()
	res := txn.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("error inserting order: %v", err)
	}

	if err = txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("error adding order to set: %v", err)
	}

	if _, err = txn.Exec(ctx); err != nil {
		return fmt.Errorf("error executing transaction: %v", err)
	}

	return nil
}

var ErrorNotExists = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)
	value, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrorNotExists
	} else if err != nil {
		return model.Order{}, fmt.Errorf("error finding order: %v", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("error unmarshalling order: %v", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)
	txn := r.Client.TxPipeline()
	err := txn.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrorNotExists
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("error deleting order: %v", err)
	}

	if err = txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("error removing order from set: %v", err)
	}

	if _, err = txn.Exec(ctx); err != nil {
		return fmt.Errorf("error executing transaction: %v", err)
	}

	return nil
}

func (r *RedisRepo) UpdateByID(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("error marshalling order: %v", err)
	}

	key := orderIDKey(order.OrderID)
	err = r.Client.SetXX(ctx, key, string(data), 0).Err()

	if errors.Is(err, redis.Nil) {
		return ErrorNotExists
	} else if err != nil {
		return fmt.Errorf("error updating order: %v", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	txn := r.Client
	res := txn.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("error scanning orders: %v", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := txn.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("error getting orders: %v", err)
	}

	orders := make([]model.Order, len(xs))
	for i, x := range xs {
		x := x.(string)
		var order model.Order
		err := json.Unmarshal([]byte(x), &order)

		if err != nil {
			return FindResult{}, fmt.Errorf("error unmarshalling order: %v", err)
		}

		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
