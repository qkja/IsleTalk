package persistence

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
)

// Connect 建立 MongoDB 连接并返回 message 独占 database（宪法原则 V 数据服务独占）。
func Connect(ctx context.Context, uri, database string) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client.Database(database), nil
}

// —— MessageRepository（T006，集合 message.messages）——

type messageDoc struct {
	MessageID      string `bson:"_id"` // 幂等键，_id 天然唯一
	ConversationID string `bson:"conversation_id"`
	SenderID       string `bson:"sender_id"`
	Type           string `bson:"type"`
	Content        string `bson:"content"`
	Seq            uint64 `bson:"seq"`
	SentAt         int64  `bson:"sent_at"`
	Status         string `bson:"status"`
	Recalled       bool   `bson:"recalled"`
}

type MongoMessageRepo struct{ col *mongo.Collection }

func NewMongoMessageRepo(db *mongo.Database) *MongoMessageRepo {
	return &MongoMessageRepo{col: db.Collection("message.messages")}
}

// EnsureIndexes 建立索引：_id（message_id）唯一（天然）+ (conversation_id, seq) 唯一（时序 FR-031）。
func (r *MongoMessageRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "conversation_id", Value: 1}, {Key: "seq", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoMessageRepo) Save(ctx context.Context, m entity.Message) (entity.Message, error) {
	_, err := r.col.InsertOne(ctx, messageDoc{
		MessageID:      m.MessageID().String(),
		ConversationID: m.ConversationID(),
		SenderID:       m.SenderID().String(),
		Type:           string(m.Type()),
		Content:        m.Content().String(),
		Seq:            m.Seq(),
		SentAt:         m.SentAt().UnixMilli(),
		Status:         string(m.Status()),
		Recalled:       m.Recalled(),
	})
	return m, err
}

func (r *MongoMessageRepo) FindByID(ctx context.Context, id valueobject.MessageId) (entity.Message, error) {
	var d messageDoc
	err := r.col.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Message{}, port.ErrNotFound
		}
		return entity.Message{}, err
	}
	return fromMessageDoc(d), nil
}

func (r *MongoMessageRepo) ListByConversation(ctx context.Context, conversationID string, beforeSeq uint64, limit int) ([]entity.Message, error) {
	filter := bson.M{"conversation_id": conversationID}
	if beforeSeq > 0 {
		filter["seq"] = bson.M{"$lt": beforeSeq}
	}
	opts := options.Find().SetSort(bson.D{{Key: "seq", Value: -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []entity.Message{}
	for cur.Next(ctx) {
		var d messageDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		out = append(out, fromMessageDoc(d))
	}
	return out, cur.Err()
}

func (r *MongoMessageRepo) MarkRead(ctx context.Context, conversationID string, readerID valueobject.UserId, upToSeq uint64) (int64, error) {
	res, err := r.col.UpdateMany(ctx,
		bson.M{
			"conversation_id": conversationID,
			"seq":             bson.M{"$lte": upToSeq},
			"sender_id":       bson.M{"$ne": readerID.String()},
			"status":          string(entity.MessageDelivered),
		},
		bson.M{"$set": bson.M{"status": string(entity.MessageRead)}},
	)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (r *MongoMessageRepo) Recall(ctx context.Context, id valueobject.MessageId) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id.String()}, bson.M{"$set": bson.M{"recalled": true}})
	return err
}

func fromMessageDoc(d messageDoc) entity.Message {
	mid, _ := valueobject.NewMessageId(d.MessageID)
	sid, _ := valueobject.NewUserId(d.SenderID)
	content, _ := valueobject.NewMessageContent(d.Content)
	return entity.ReconstructMessage(
		mid, d.ConversationID, sid, entity.MessageType(d.Type), content,
		d.Seq, time.UnixMilli(d.SentAt), entity.MessageStatus(d.Status), d.Recalled,
	)
}

// —— ConversationRepository（T006，集合 message.conversations）——

type conversationDoc struct {
	ConversationID     string   `bson:"_id"` // conversation_id 天然唯一
	Type               string   `bson:"type"`
	Participants       []string `bson:"participants"`
	LastMessagePreview string   `bson:"last_message_preview"`
	UnreadCount        int      `bson:"unread_count"`
	Pinned             bool     `bson:"pinned"`
	Muted              bool     `bson:"muted"`
	Deleted            bool     `bson:"deleted"`
	LastActiveAt       int64    `bson:"last_active_at"`
	Seq                uint64   `bson:"seq"` // 内部原子序号计数器（非实体字段，FR-031）
}

type MongoConversationRepo struct{ col *mongo.Collection }

func NewMongoConversationRepo(db *mongo.Database) *MongoConversationRepo {
	return &MongoConversationRepo{col: db.Collection("message.conversations")}
}

// EnsureIndexes 建立索引：_id（conversation_id）唯一（天然）+ participants 普通（列表查询）。
func (r *MongoConversationRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "participants", Value: 1}},
	})
	return err
}

func (r *MongoConversationRepo) Save(ctx context.Context, c entity.Conversation) (entity.Conversation, error) {
	parts := make([]string, 0, len(c.Participants()))
	for _, p := range c.Participants() {
		parts = append(parts, p.String())
	}
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": c.ConversationID()},
		bson.M{"$set": bson.M{
			"type":                 string(c.Type()),
			"participants":         parts,
			"last_message_preview": c.LastMessagePreview(),
			"unread_count":         c.UnreadCount(),
			"pinned":               c.Pinned(),
			"muted":                c.Muted(),
			"deleted":              c.Deleted(),
			"last_active_at":       c.LastActiveAt().UnixMilli(),
		}, "$setOnInsert": bson.M{"_id": c.ConversationID(), "seq": 0}},
		options.Update().SetUpsert(true))
	return c, err
}

func (r *MongoConversationRepo) FindByID(ctx context.Context, conversationID string) (entity.Conversation, error) {
	var d conversationDoc
	err := r.col.FindOne(ctx, bson.M{"_id": conversationID}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Conversation{}, port.ErrNotFound
		}
		return entity.Conversation{}, err
	}
	return fromConversationDoc(d), nil
}

func (r *MongoConversationRepo) ListByUser(ctx context.Context, userID valueobject.UserId, limit int) ([]entity.Conversation, error) {
	// 置顶优先，其余按最近活跃倒序（FR-008）。
	opts := options.Find().SetSort(bson.D{{Key: "pinned", Value: -1}, {Key: "last_active_at", Value: -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, bson.M{"participants": userID.String()}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []entity.Conversation{}
	for cur.Next(ctx) {
		var d conversationDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		out = append(out, fromConversationDoc(d))
	}
	return out, cur.Err()
}

// NextSeq 原子分配会话内下一个 seq（FR-031 单调递增）。
func (r *MongoConversationRepo) NextSeq(ctx context.Context, conversationID string) (uint64, error) {
	res := r.col.FindOneAndUpdate(ctx,
		bson.M{"_id": conversationID},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var d conversationDoc
	if err := res.Decode(&d); err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, port.ErrNotFound
		}
		return 0, err
	}
	return d.Seq, nil
}

// TouchOnSend 消息发送后原子更新摘要/未读/活跃时间，并恢复已删除会话（FR-008/024）。
func (r *MongoConversationRepo) TouchOnSend(ctx context.Context, conversationID, preview string) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": conversationID},
		bson.M{
			"$inc": bson.M{"unread_count": 1},
			"$set": bson.M{"last_message_preview": preview, "last_active_at": time.Now().UnixMilli(), "deleted": false},
		},
	)
	return err
}

// ResetUnread 未读清零（FR-019 已读上报）。
func (r *MongoConversationRepo) ResetUnread(ctx context.Context, conversationID string) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": conversationID}, bson.M{"$set": bson.M{"unread_count": 0}})
	return err
}

// TouchOnRecall 撤回后更新会话「最新消息摘要」为占位（FR-006，不增未读）。
func (r *MongoConversationRepo) TouchOnRecall(ctx context.Context, conversationID, preview string) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": conversationID}, bson.M{"$set": bson.M{"last_message_preview": preview}})
	return err
}

func fromConversationDoc(d conversationDoc) entity.Conversation {
	parts := make([]valueobject.UserId, 0, len(d.Participants))
	for _, p := range d.Participants {
		if u, err := valueobject.NewUserId(p); err == nil {
			parts = append(parts, u)
		}
	}
	return entity.ReconstructConversation(
		d.ConversationID, entity.ConversationType(d.Type), parts, d.LastMessagePreview,
		d.UnreadCount, d.Pinned, d.Muted, d.Deleted, time.UnixMilli(d.LastActiveAt),
	)
}
