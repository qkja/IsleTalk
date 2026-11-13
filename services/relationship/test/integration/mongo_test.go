package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/valueobject"
	"isletalk/relationship/infrastructure/persistence"
)

// 集成测试（T033）：针对容器化 MongoDB。本地无容器时 skip。

func mongoDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Skipf("MongoDB 不可用: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB ping 失败: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	db := client.Database("relationship_test")
	_ = db.Drop(context.Background()) // 清理旧数据，保证测试隔离
	return db
}

func mustUID(s string) valueobject.UserId {
	v, _ := valueobject.NewUserId(s)
	return v
}

func TestMongoFriendshipBidirectional(t *testing.T) {
	db := mongoDB(t)
	repo := persistence.NewMongoFriendshipRepo(db)
	ctx := context.Background()
	_ = repo.EnsureIndexes(ctx)

	a, b := mustUID("a"), mustUID("b")
	if _, err := repo.Create(ctx, entity.NewFriendship(a, b)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, entity.NewFriendship(b, a)); err != nil {
		t.Fatal(err)
	}

	if f, err := repo.Find(ctx, a, b); err != nil || !f.IsActive() {
		t.Fatal("friendship A→B should be active")
	}
	if f, err := repo.Find(ctx, b, a); err != nil || !f.IsActive() {
		t.Fatal("friendship B→A should be active")
	}
}

func TestMongoFriendRequestUnique(t *testing.T) {
	db := mongoDB(t)
	repo := persistence.NewMongoFriendRequestRepo(db)
	ctx := context.Background()
	_ = repo.EnsureIndexes(ctx)

	a, b := mustUID("a"), mustUID("b")
	if _, err := repo.Create(ctx, entity.NewFriendRequest(a, b, "hi")); err != nil {
		t.Fatal(err)
	}
	// 重复申请应因唯一索引失败
	if _, err := repo.Create(ctx, entity.NewFriendRequest(a, b, "again")); err == nil {
		t.Fatal("duplicate request should fail on unique index")
	}
}

func TestMongoBlockSet(t *testing.T) {
	db := mongoDB(t)
	repo := persistence.NewMongoBlockRepo(db)
	ctx := context.Background()
	_ = repo.EnsureIndexes(ctx)

	a, b := mustUID("a"), mustUID("b")
	if err := repo.Set(ctx, a, b, true); err != nil {
		t.Fatal(err)
	}
	if blk, err := repo.Find(ctx, a, b); err != nil || !blk.IsActive() {
		t.Fatal("block should be active")
	}
}
