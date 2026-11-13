package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
	"isletalk/message/infrastructure/persistence"
)

func TestMongoConnect(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db, err := persistence.Connect(ctx, uri, "message_test")
	if err != nil {
		t.Skipf("MongoDB 不可用: %v", err)
	}
	_ = db.Drop(context.Background())
}

func TestMongoListMessages(t *testing.T) {
	db := mongoDB(t)
	repo := persistence.NewMongoMessageRepo(db)
	ctx := context.Background()
	_ = repo.EnsureIndexes(ctx)

	for i := 1; i <= 3; i++ {
		m := entity.NewMessage(mustMID(fmt.Sprintf("m%d", i)), "direct:a:b", mustUID("a"), entity.MessageText, mustContent("x"), uint64(i))
		if _, err := repo.Save(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.ListByConversation(ctx, "direct:a:b", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Seq() != 3 {
		t.Fatalf("list len = %d, first seq = %d; want 3 and 3 (descending)", len(got), got[0].Seq())
	}
}

func TestMongoListConversationsByUser(t *testing.T) {
	db := mongoDB(t)
	repo := persistence.NewMongoConversationRepo(db)
	ctx := context.Background()
	_ = repo.EnsureIndexes(ctx)

	conv := entity.NewConversation("direct:a:b", []valueobject.UserId{mustUID("a"), mustUID("b")})
	if _, err := repo.Save(ctx, conv); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ListByUser(ctx, mustUID("a"), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("list len = %d, want 1", len(got))
	}
}
