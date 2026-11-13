package persistence

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// 4 个仓储 port（T007），领域模型与 BSON 解耦（宪法原则 V）。

// —— FriendshipRepository ——

type friendshipDoc struct {
	UserID    string `bson:"_id"`
	FriendID  string `bson:"friend_id"`
	CreatedAt int64  `bson:"created_at"`
	Status    string `bson:"status"`
}

type MongoFriendshipRepo struct{ col *mongo.Collection }

func NewMongoFriendshipRepo(db *mongo.Database) *MongoFriendshipRepo {
	return &MongoFriendshipRepo{col: db.Collection("relationship.friendships")}
}

// EnsureIndexes 建立唯一索引：(user_id, friend_id) 唯一。
func (r *MongoFriendshipRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}, {Key: "friend_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoFriendshipRepo) Create(ctx context.Context, f entity.Friendship) (entity.Friendship, error) {
	_, err := r.col.InsertOne(ctx, friendshipDoc{
		UserID:    f.UserID().String(),
		FriendID:  f.FriendID().String(),
		CreatedAt: f.CreatedAt().UnixMilli(),
		Status:    string(f.Status()),
	})
	return f, err
}

func (r *MongoFriendshipRepo) Find(ctx context.Context, userID, friendID valueobject.UserId) (entity.Friendship, error) {
	var d friendshipDoc
	err := r.col.FindOne(ctx, bson.M{"_id": userID.String(), "friend_id": friendID.String()}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Friendship{}, port.ErrNotFound
		}
		return entity.Friendship{}, err
	}
	uid, _ := valueobject.NewUserId(d.UserID)
	fid, _ := valueobject.NewUserId(d.FriendID)
	f := entity.NewFriendship(uid, fid)
	if d.Status == string(entity.FriendshipDeleted) {
		f.Delete()
	}
	return f, nil
}

func (r *MongoFriendshipRepo) ListByUser(ctx context.Context, userID valueobject.UserId) ([]entity.Friendship, error) {
	cur, err := r.col.Find(ctx, bson.M{"_id": userID.String()})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []entity.Friendship{}
	for cur.Next(ctx) {
		var d friendshipDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		fid, _ := valueobject.NewUserId(d.FriendID)
		out = append(out, entity.NewFriendship(userID, fid))
	}
	return out, cur.Err()
}

func (r *MongoFriendshipRepo) Delete(ctx context.Context, userID, friendID valueobject.UserId) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": userID.String(), "friend_id": friendID.String()},
		bson.M{"$set": bson.M{"status": "deleted"}})
	return err
}

// —— FriendRequestRepository ——

type friendRequestDoc struct {
	ID             string `bson:"_id"`
	InitiatorID    string `bson:"initiator_id"`
	TargetID       string `bson:"target_id"`
	RequestMessage string `bson:"request_message"`
	Status         string `bson:"status"`
	CreatedAt      int64  `bson:"created_at"`
	RespondedAt    int64  `bson:"responded_at"`
}

type MongoFriendRequestRepo struct{ col *mongo.Collection }

func NewMongoFriendRequestRepo(db *mongo.Database) *MongoFriendRequestRepo {
	return &MongoFriendRequestRepo{col: db.Collection("relationship.friend_requests")}
}

// EnsureIndexes 建立唯一索引：(initiator_id, target_id) 唯一。
func (r *MongoFriendRequestRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "initiator_id", Value: 1}, {Key: "target_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoFriendRequestRepo) Create(ctx context.Context, req entity.FriendRequest) (entity.FriendRequest, error) {
	_, err := r.col.InsertOne(ctx, friendRequestDoc{
		ID:             req.InitiatorID().String() + ":" + req.TargetID().String(),
		InitiatorID:    req.InitiatorID().String(),
		TargetID:       req.TargetID().String(),
		RequestMessage: req.RequestMessage(),
		Status:         string(req.Status()),
		CreatedAt:      req.CreatedAt().UnixMilli(),
	})
	return req, err
}

func (r *MongoFriendRequestRepo) Find(ctx context.Context, initiatorID, targetID valueobject.UserId) (entity.FriendRequest, error) {
	var d friendRequestDoc
	err := r.col.FindOne(ctx, bson.M{"initiator_id": initiatorID.String(), "target_id": targetID.String()}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.FriendRequest{}, port.ErrNotFound
		}
		return entity.FriendRequest{}, err
	}
	return fromFriendRequestDoc(d), nil
}

func (r *MongoFriendRequestRepo) ListPendingByTarget(ctx context.Context, targetID valueobject.UserId) ([]entity.FriendRequest, error) {
	return r.list(ctx, bson.M{"target_id": targetID.String(), "status": "pending"})
}

func (r *MongoFriendRequestRepo) ListHistoryByTarget(ctx context.Context, targetID valueobject.UserId) ([]entity.FriendRequest, error) {
	return r.list(ctx, bson.M{"target_id": targetID.String(), "status": bson.M{"$ne": "pending"}})
}

func (r *MongoFriendRequestRepo) list(ctx context.Context, filter bson.M) ([]entity.FriendRequest, error) {
	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []entity.FriendRequest{}
	for cur.Next(ctx) {
		var d friendRequestDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		out = append(out, fromFriendRequestDoc(d))
	}
	return out, cur.Err()
}

func fromFriendRequestDoc(d friendRequestDoc) entity.FriendRequest {
	iid, _ := valueobject.NewUserId(d.InitiatorID)
	tid, _ := valueobject.NewUserId(d.TargetID)
	r := entity.NewFriendRequest(iid, tid, d.RequestMessage)
	switch entity.FriendRequestStatus(d.Status) {
	case entity.RequestAccepted:
		r.Accept()
	case entity.RequestRejected:
		r.Reject()
	}
	return r
}

func (r *MongoFriendRequestRepo) Update(ctx context.Context, req entity.FriendRequest) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"initiator_id": req.InitiatorID().String(), "target_id": req.TargetID().String()},
		bson.M{"$set": bson.M{"status": string(req.Status())}})
	return err
}

// —— BlockRepository ——

type blockDoc struct {
	OwnerID  string `bson:"_id"`
	TargetID string `bson:"target_id"`
	Status   string `bson:"status"`
}

type MongoBlockRepo struct{ col *mongo.Collection }

func NewMongoBlockRepo(db *mongo.Database) *MongoBlockRepo {
	return &MongoBlockRepo{col: db.Collection("relationship.blocks")}
}

func (r *MongoBlockRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}, {Key: "target_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoBlockRepo) Set(ctx context.Context, ownerID, targetID valueobject.UserId, blocked bool) error {
	status := entity.BlockInactive
	if blocked {
		status = entity.BlockActive
	}
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": ownerID.String(), "target_id": targetID.String()},
		bson.M{"$set": bson.M{"status": string(status)}, "$setOnInsert": bson.M{"_id": ownerID.String(), "target_id": targetID.String()}},
		options.Update().SetUpsert(true))
	return err
}

func (r *MongoBlockRepo) Find(ctx context.Context, ownerID, targetID valueobject.UserId) (entity.Block, error) {
	var d blockDoc
	err := r.col.FindOne(ctx, bson.M{"_id": ownerID.String(), "target_id": targetID.String()}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Block{}, port.ErrNotFound
		}
		return entity.Block{}, err
	}
	oid, _ := valueobject.NewUserId(d.OwnerID)
	tid, _ := valueobject.NewUserId(d.TargetID)
	b := entity.NewBlock(oid, tid)
	if d.Status == string(entity.BlockInactive) {
		b.Unblock()
	}
	return b, nil
}

// —— RemarkRepository ——

type remarkDoc struct {
	OwnerID   string `bson:"_id"`
	FriendID  string `bson:"friend_id"`
	Content   string `bson:"content"`
	UpdatedAt int64  `bson:"updated_at"`
}

type MongoRemarkRepo struct{ col *mongo.Collection }

func NewMongoRemarkRepo(db *mongo.Database) *MongoRemarkRepo {
	return &MongoRemarkRepo{col: db.Collection("relationship.remarks")}
}

func (r *MongoRemarkRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}, {Key: "friend_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoRemarkRepo) Save(ctx context.Context, remark entity.Remark) (entity.Remark, error) {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": remark.OwnerID().String(), "friend_id": remark.FriendID().String()},
		bson.M{"$set": bson.M{"content": remark.Content().String(), "updated_at": time.Now().UnixMilli()},
			"$setOnInsert": bson.M{"_id": remark.OwnerID().String(), "friend_id": remark.FriendID().String()}},
		options.Update().SetUpsert(true))
	return remark, err
}

func (r *MongoRemarkRepo) Find(ctx context.Context, ownerID, friendID valueobject.UserId) (entity.Remark, error) {
	var d remarkDoc
	err := r.col.FindOne(ctx, bson.M{"_id": ownerID.String(), "friend_id": friendID.String()}).Decode(&d)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Remark{}, port.ErrNotFound
		}
		return entity.Remark{}, err
	}
	txt, _ := valueobject.NewRemarkText(d.Content)
	oid, _ := valueobject.NewUserId(d.OwnerID)
	fid, _ := valueobject.NewUserId(d.FriendID)
	return entity.NewRemark(oid, fid, txt), nil
}
