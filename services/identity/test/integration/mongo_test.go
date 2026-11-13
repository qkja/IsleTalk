package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"isletalk/identity/domain/entity"
	"isletalk/identity/infrastructure/persistence"
)

// 集成测试（T042）：针对容器化 MongoDB 运行。本地无容器时 skip。
func mongoTestDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Skipf("MongoDB 不可用，跳过集成测试: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB ping 失败，跳过集成测试: %v", err)
	}
	db := client.Database("identity_test")
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})
	return db
}

func testUser(t *testing.T, id, name, phone string) entity.User {
	t.Helper()
	uid, _ := newID(id)
	pn, _ := newName(name)
	ph, _ := newPhone("phone:" + phone)
	u, err := entity.NewUser(uid, pn, ph, "zh-CN")
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	return u
}

func TestMongoUserRepoCreateFind(t *testing.T) {
	db := mongoTestDB(t)
	ctx := context.Background()
	repo := persistence.NewMongoUserRepo(db)
	if err := repo.EnsureIndexes(ctx); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}

	u := testUser(t, "u-1", "alice", "138")
	if _, err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.FindByPhoneHash(ctx, u.PhoneHash())
	if err != nil {
		t.Fatalf("FindByPhoneHash: %v", err)
	}
	if got.PublicName().String() != "alice" {
		t.Fatalf("public name = %q", got.PublicName())
	}

	byName, err := repo.FindByPublicName(ctx, u.PublicName())
	if err != nil || byName.ID() != u.ID() {
		t.Fatalf("FindByPublicName = %v, %v", byName.ID(), err)
	}
}

func TestMongoUserRepoUniquePhoneHash(t *testing.T) {
	db := mongoTestDB(t)
	ctx := context.Background()
	repo := persistence.NewMongoUserRepo(db)
	_ = repo.EnsureIndexes(ctx)

	u1 := testUser(t, "u-1", "alice", "138")
	u2 := testUser(t, "u-2", "bob", "138") // 同手机号
	if _, err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("Create u1: %v", err)
	}
	if _, err := repo.Create(ctx, u2); err == nil {
		t.Fatal("duplicate phone_hash should be rejected by unique index")
	}
}

func TestMongoSessionRepoRevoke(t *testing.T) {
	db := mongoTestDB(t)
	ctx := context.Background()
	repo := persistence.NewMongoSessionRepo(db)

	uid, _ := newID("u-1")
	sess := entity.NewSession("sess-1", uid, "dev-1")
	if _, err := repo.Create(ctx, sess); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Revoke(ctx, "sess-1"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	got, err := repo.FindByID(ctx, "sess-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if !got.Revoked() {
		t.Fatal("session should be revoked")
	}
}

func TestMongoNameClaimAvailability(t *testing.T) {
	db := mongoTestDB(t)
	ctx := context.Background()
	repo := persistence.NewMongoNameClaimRepo(db)
	_ = repo.EnsureIndexes(ctx)

	uid, _ := newID("u-1")
	name, _ := newName("alice")

	ok, err := repo.IsAvailable(ctx, name)
	if err != nil || !ok {
		t.Fatalf("IsAvailable(empty) = %v, %v", ok, err)
	}

	if err := repo.Claim(ctx, name, uid); err != nil {
		t.Fatalf("Claim: %v", err)
	}
	ok, _ = repo.IsAvailable(ctx, name)
	if ok {
		t.Fatal("claimed name should not be available")
	}

	// 释放后进入保留期（90 天），仍不可用
	if err := repo.Release(ctx, name); err != nil {
		t.Fatalf("Release: %v", err)
	}
	ok, _ = repo.IsAvailable(ctx, name)
	if ok {
		t.Fatal("released name within retention should not be available")
	}
}

func TestMongoVerificationSaveFindRevoke(t *testing.T) {
	db := mongoTestDB(t)
	ctx := context.Background()
	repo := persistence.NewMongoVerificationRepo(db)

	uid, _ := newID("u-1")
	v := entity.NewVerification(uid, "idcard", "encrypted-digest", time.Now(), time.Now().Add(30*24*time.Hour))
	saved, err := repo.Save(ctx, v)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.Status() != entity.VerificationSubmitted {
		t.Fatalf("status = %v", saved.Status())
	}

	got, err := repo.FindByUser(ctx, uid)
	if err != nil || got.IDType() != "idcard" {
		t.Fatalf("FindByUser = %q, %v", got.IDType(), err)
	}

	if err := repo.Revoke(ctx, uid); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	revoked, _ := repo.FindByUser(ctx, uid)
	if revoked.Status() != entity.VerificationRevoked {
		t.Fatal("verification should be revoked")
	}
}

func TestMongoUserFindByIDAndUpdate(t *testing.T) {
	db := mongoTestDB(t)
	ctx := context.Background()
	repo := persistence.NewMongoUserRepo(db)
	_ = repo.EnsureIndexes(ctx)

	u := testUser(t, "u-1", "alice", "138")
	if _, err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.FindByID(ctx, u.ID())
	if err != nil || got.PublicName().String() != "alice" {
		t.Fatalf("FindByID: %v", err)
	}

	got.UpdateProfile("Alice", "", "")
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	updated, _ := repo.FindByID(ctx, u.ID())
	if updated.Nickname() != "Alice" {
		t.Fatal("nickname should be updated")
	}
}

func TestMongoSessionListRevokeAll(t *testing.T) {
	db := mongoTestDB(t)
	ctx := context.Background()
	repo := persistence.NewMongoSessionRepo(db)

	uid, _ := newID("u-1")
	if _, err := repo.Create(ctx, entity.NewSession("s-1", uid, "d1")); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, entity.NewSession("s-2", uid, "d2")); err != nil {
		t.Fatal(err)
	}

	list, err := repo.ListByUser(ctx, uid)
	if err != nil || len(list) != 2 {
		t.Fatalf("ListByUser len = %d, %v", len(list), err)
	}

	if err := repo.Touch(ctx, "s-1"); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	if err := repo.RevokeAllByUser(ctx, uid); err != nil {
		t.Fatalf("RevokeAllByUser: %v", err)
	}
	list2, _ := repo.ListByUser(ctx, uid)
	for _, s := range list2 {
		if !s.Revoked() {
			t.Fatal("all sessions should be revoked")
		}
	}
}
