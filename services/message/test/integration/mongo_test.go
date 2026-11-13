package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
	"isletalk/message/infrastructure/persistence"
)

// 集成测试（T026）：针对容器化 MongoDB。本地无容器时 skip。

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
	db := client.Database("message_test")
	_ = db.Drop(context.Background())
	return db
}

func mustUID(s string) valueobject.UserId    { v, _ := valueobject.NewUserId(s); return v }
func mustMID(s string) valueobject.MessageId { v, _ := valueobject.NewMessageId(s); return v }
func mustContent(s string) valueobject.MessageContent {
	v, _ := valueobject.NewMessageContent(s)
	return v
}

func TestMongoMessageIdempotentUnique(t *testing.T) {
	db := mongoDB(t)
	repo := persistence.NewMongoMessageRepo(db)
	ctx := context.Background()
	_ = repo.EnsureIndexes(ctx)

	m := entity.NewMessage(mustMID("m1"), "direct:a:b", mustUID("a"), entity.MessageText, mustContent("hi"), 1)
	if _, err := repo.Save(ctx, m); err != nil {
		t.Fatal(err)
	}
	// 相同 message_id（_id 唯一）应失败
	if _, err := repo.Save(ctx, m); err == nil {
		t.Fatal("duplicate message_id should fail on unique index")
	}
}

func TestMongoSeqMonotonicUnique(t *testing.T) {
	db := mongoDB(t)
	mrepo := persistence.NewMongoMessageRepo(db)
	crepo := persistence.NewMongoConversationRepo(db)
	ctx := context.Background()
	_ = mrepo.EnsureIndexes(ctx)
	_ = crepo.EnsureIndexes(ctx)

	conv := entity.NewConversation("direct:a:b", []valueobject.UserId{mustUID("a"), mustUID("b")})
	if _, err := crepo.Save(ctx, conv); err != nil {
		t.Fatal(err)
	}

	seq1, err := crepo.NextSeq(ctx, "direct:a:b")
	if err != nil {
		t.Fatal(err)
	}
	seq2, err := crepo.NextSeq(ctx, "direct:a:b")
	if err != nil {
		t.Fatal(err)
	}
	if seq1 != 1 || seq2 != 2 {
		t.Fatalf("seq = %d, %d, want 1, 2 (monotonic)", seq1, seq2)
	}

	m := entity.NewMessage(mustMID("m1"), "direct:a:b", mustUID("a"), entity.MessageText, mustContent("hi"), 1)
	if _, err := mrepo.Save(ctx, m); err != nil {
		t.Fatal(err)
	}
	dup := entity.NewMessage(mustMID("m2"), "direct:a:b", mustUID("a"), entity.MessageText, mustContent("hi"), 1)
	if _, err := mrepo.Save(ctx, dup); err == nil {
		t.Fatal("duplicate (conversation_id, seq) should fail on unique index")
	}
}

func TestMongoUnreadAndRecall(t *testing.T) {
	db := mongoDB(t)
	mrepo := persistence.NewMongoMessageRepo(db)
	crepo := persistence.NewMongoConversationRepo(db)
	ctx := context.Background()
	_ = mrepo.EnsureIndexes(ctx)
	_ = crepo.EnsureIndexes(ctx)

	conv := entity.NewConversation("direct:a:b", []valueobject.UserId{mustUID("a"), mustUID("b")})
	if _, err := crepo.Save(ctx, conv); err != nil {
		t.Fatal(err)
	}

	seq, _ := crepo.NextSeq(ctx, "direct:a:b")
	m := entity.NewMessage(mustMID("m1"), "direct:a:b", mustUID("a"), entity.MessageText, mustContent("hi"), seq)
	m.MarkDelivered()
	if _, err := mrepo.Save(ctx, m); err != nil {
		t.Fatal(err)
	}

	// 发送后未读 +1（免打扰仍计入未读）
	if err := crepo.TouchOnSend(ctx, "direct:a:b", "hi"); err != nil {
		t.Fatal(err)
	}
	c, err := crepo.FindByID(ctx, "direct:a:b")
	if err != nil {
		t.Fatal(err)
	}
	if c.UnreadCount() != 1 {
		t.Fatalf("unread = %d, want 1", c.UnreadCount())
	}

	// 已读上报：消息 read + 未读清零
	if _, err := mrepo.MarkRead(ctx, "direct:a:b", mustUID("b"), seq); err != nil {
		t.Fatal(err)
	}
	got, err := mrepo.FindByID(ctx, mustMID("m1"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Status() != entity.MessageRead {
		t.Fatalf("status = %v, want read", got.Status())
	}
	if err := crepo.ResetUnread(ctx, "direct:a:b"); err != nil {
		t.Fatal(err)
	}
	c, _ = crepo.FindByID(ctx, "direct:a:b")
	if c.UnreadCount() != 0 {
		t.Fatalf("unread = %d, want 0", c.UnreadCount())
	}

	// 撤回
	if err := mrepo.Recall(ctx, mustMID("m1")); err != nil {
		t.Fatal(err)
	}
	got, _ = mrepo.FindByID(ctx, mustMID("m1"))
	if !got.Recalled() {
		t.Fatal("message should be recalled")
	}
}
