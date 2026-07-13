package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/pkg/jsonhelper"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var _ CommonBehaviour[entity.Entity] = &RedisCommonBehaviour[entity.Entity]{}

type RedisCommonBehaviour[T entity.Entity] struct {
	client rueidis.Client
}

func NewRedisCommonBehaviour[T entity.Entity](client rueidis.Client) *RedisCommonBehaviour[T] {
	return &RedisCommonBehaviour[T]{
		client: client,
	}
}

func (r RedisCommonBehaviour[T]) Get(ctx context.Context, id entity.ID) (T, error) {
	var t T
	cmd := r.client.B().JsonGet().Key(id.String()).Path(".").Build()
	val, err := r.client.Do(ctx, cmd).ToString()
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			return t, ErrNotFound
		}
		logrus.WithError(err).WithField("id", id).Errorln("failed to get from redis")
		return t, err
	}
	return jsonhelper.Decode[T]([]byte(val)), nil
}

func (r RedisCommonBehaviour[T]) Save(ctx context.Context, t T) error {
	cmd := r.client.B().JsonSet().Key(t.EntityID().String()).Path("$").Value(string(jsonhelper.Encode[T](t))).Build()
	if err := r.client.Do(ctx, cmd).Error(); err != nil {
		logrus.WithError(err).WithField("entity", t).Errorln("failed to save entity")
		return err
	}
	return nil
}

func (r RedisCommonBehaviour[T]) MGet(ctx context.Context, IDs ...entity.ID) ([]T, error) {
	keys := lo.Map(IDs, func(ID entity.ID, _ int) string {
		return ID.String()
	})
	cmd := r.client.B().JsonMget().Key(keys...).Path(".").Build()
	vals, err := r.client.Do(ctx, cmd).AsStrSlice()
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			return nil, ErrNotFound
		}
		logrus.WithError(err).WithField("IDs", IDs).Errorln("failed to get many from redis")
		return nil, err
	}
	return lo.Map(lo.Filter(vals, func(item string, _ int) bool {
		return item != ""
	}), func(item string, _ int) T {
		return jsonhelper.Decode[T]([]byte(item))
	}), nil
}

func (r RedisCommonBehaviour[T]) MSet(ctx context.Context, entities ...T) error {
	if len(entities) == 0 {
		return nil
	}
	var cmd = r.client.B().JsonMset().Key(entities[0].EntityID().String()).Path(".").
		Value(string(jsonhelper.Encode(entities[0])))

	for i, ent := range entities {
		if i == 0 {
			continue
		}
		cmd = cmd.Key(ent.EntityID().String()).Path(".").Value(string(jsonhelper.Encode(ent)))
	}

	err := r.client.Do(ctx, cmd.Build()).Error()
	if err != nil {
		logrus.WithError(err).Errorln("could not save multi items")
		return err
	}
	return nil
}

func (r RedisCommonBehaviour[T]) SetField(ctx context.Context, ID entity.ID, fieldName string, value any) error {
	cmd := r.client.B().JsonSet().Key(ID.String()).Path(fmt.Sprintf("$.%s", fieldName)).
		Value(string(jsonhelper.Encode(value))).Build()
	if err := r.client.Do(ctx, cmd).Error(); err != nil {
		logrus.WithError(err).WithField("entity", ID).Errorln("failed to udpate entity")
		return err
	}
	return nil
}
