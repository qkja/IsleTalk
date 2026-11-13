package persistence

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
	"isletalk/identity/infrastructure/config"
)

// 本文件实现 4 个仓储 port（T009），完整 CRUD。
// 领域模型与 BSON 解耦：实体私有字段经 DTO 映射（宪法原则 V）。

// —— 实体 ↔ BSON DTO ——

type userDoc struct {
	ID                     string `bson:"_id"`
	PublicName             string `bson:"public_name"`
	PhoneHash              string `bson:"phone_hash"`
	PasswordHash           string `bson:"password_hash"`
	Nickname               string `bson:"nickname"`
	AvatarURL              string `bson:"avatar_url"`
	Signature              string `bson:"signature"`
	Visibility             string `bson:"visibility"`
	AllowStrangerAdd       bool   `bson:"allow_stranger_add"`
	Locale                 string `bson:"locale"`
	Status                 string `bson:"status"`
	CreatedAt              int64  `bson:"created_at"`
	LastPublicNameChangeAt int64  `bson:"last_public_name_change_at"`
}

func toUserDoc(u entity.User) userDoc {
	return userDoc{
		ID:                     u.ID().String(),
		PublicName:             u.PublicName().String(),
		PhoneHash:              u.PhoneHash().String(),
		PasswordHash:           u.PasswordHash().String(),
		Nickname:               u.Nickname(),
		AvatarURL:              u.AvatarURL(),
		Signature:              u.Signature(),
		Visibility:             string(u.Visibility()),
		AllowStrangerAdd:       u.AllowStrangerAdd(),
		Locale:                 u.Locale(),
		Status:                 string(u.Status()),
		CreatedAt:              u.CreatedAt().UnixMilli(),
		LastPublicNameChangeAt: u.LastPublicNameChangeAt().UnixMilli(),
	}
}

func fromUserDoc(d userDoc) (entity.User, error) {
	id, err := valueobject.NewUserId(d.ID)
	if err != nil {
		return entity.User{}, err
	}
	name, err := valueobject.NewPublicName(d.PublicName)
	if err != nil {
		return entity.User{}, err
	}
	phone, err := valueobject.NewPhoneHash(d.PhoneHash)
	if err != nil {
		return entity.User{}, err
	}
	u, err := entity.NewUser(id, name, phone, d.Locale)
	if err != nil {
		return entity.User{}, err
	}
	if pw, err := valueobject.NewPasswordHash(d.PasswordHash); err == nil {
		u.SetPassword(pw)
	}
	u.UpdateProfile(d.Nickname, d.AvatarURL, d.Signature)
	u.SetVisibility(entity.Visibility(d.Visibility))
	u.SetAllowStrangerAdd(d.AllowStrangerAdd)
	if d.LastPublicNameChangeAt > 0 {
		u.RestoreLastPublicNameChangeAt(time.UnixMilli(d.LastPublicNameChangeAt))
	}
	if d.Status == string(entity.UserDeactivated) {
		u.Deactivate()
	}
	return u, nil
}

// —— UserRepository ——

type MongoUserRepo struct{ users *mongo.Collection }

func NewMongoUserRepo(db *mongo.Database) *MongoUserRepo {
	return &MongoUserRepo{users: db.Collection("identity.users")}
}

// EnsureIndexes 建立唯一索引：phone_hash 唯一、public_name 部分唯一（仅 active）。
func (r *MongoUserRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.users.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "phone_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
		{
			Keys:    bson.D{{Key: "public_name", Value: 1}},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"status": "active"}),
		},
	})
	return err
}

func (r *MongoUserRepo) Create(ctx context.Context, u entity.User) (entity.User, error) {
	_, err := r.users.InsertOne(ctx, toUserDoc(u))
	if err != nil {
		return entity.User{}, err
	}
	return u, nil
}

func (r *MongoUserRepo) findOne(ctx context.Context, filter bson.M) (entity.User, error) {
	var d userDoc
	if err := r.users.FindOne(ctx, filter).Decode(&d); err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.User{}, port.ErrNotFound
		}
		return entity.User{}, err
	}
	return fromUserDoc(d)
}

func (r *MongoUserRepo) FindByID(ctx context.Context, id valueobject.UserId) (entity.User, error) {
	return r.findOne(ctx, bson.M{"_id": id.String()})
}
func (r *MongoUserRepo) FindByPhoneHash(ctx context.Context, h valueobject.PhoneHash) (entity.User, error) {
	return r.findOne(ctx, bson.M{"phone_hash": h.String()})
}
func (r *MongoUserRepo) FindByPublicName(ctx context.Context, n valueobject.PublicName) (entity.User, error) {
	return r.findOne(ctx, bson.M{"public_name": n.String()})
}
func (r *MongoUserRepo) Update(ctx context.Context, u entity.User) error {
	_, err := r.users.UpdateOne(
		ctx,
		bson.M{"_id": u.ID().String()},
		bson.M{"$set": toUserDoc(u)},
	)
	return err
}

// —— SessionRepository ——

type sessionDoc struct {
	ID           string `bson:"_id"`
	UserID       string `bson:"user_id"`
	DeviceID     string `bson:"device_id"`
	CreatedAt    int64  `bson:"created_at"`
	LastActiveAt int64  `bson:"last_active_at"`
	Revoked      bool   `bson:"revoked"`
}

func toSessionDoc(s entity.Session) sessionDoc {
	return sessionDoc{
		ID:           s.ID(),
		UserID:       s.UserID().String(),
		DeviceID:     s.DeviceID(),
		CreatedAt:    s.CreatedAt().UnixMilli(),
		LastActiveAt: s.LastActiveAt().UnixMilli(),
		Revoked:      s.Revoked(),
	}
}

func fromSessionDoc(d sessionDoc) entity.Session {
	uid, _ := valueobject.NewUserId(d.UserID)
	s := entity.NewSession(d.ID, uid, d.DeviceID)
	if d.Revoked {
		s.Revoke()
	}
	return s
}

type MongoSessionRepo struct{ sessions *mongo.Collection }

func NewMongoSessionRepo(db *mongo.Database) *MongoSessionRepo {
	return &MongoSessionRepo{sessions: db.Collection("identity.sessions")}
}

func (r *MongoSessionRepo) Create(ctx context.Context, s entity.Session) (entity.Session, error) {
	_, err := r.sessions.InsertOne(ctx, toSessionDoc(s))
	return s, err
}
func (r *MongoSessionRepo) FindByID(ctx context.Context, id string) (entity.Session, error) {
	var d sessionDoc
	if err := r.sessions.FindOne(ctx, bson.M{"_id": id}).Decode(&d); err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Session{}, port.ErrNotFound
		}
		return entity.Session{}, err
	}
	return fromSessionDoc(d), nil
}
func (r *MongoSessionRepo) ListByUser(ctx context.Context, uid valueobject.UserId) ([]entity.Session, error) {
	cur, err := r.sessions.Find(ctx, bson.M{"user_id": uid.String()})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []entity.Session{}
	for cur.Next(ctx) {
		var d sessionDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		out = append(out, fromSessionDoc(d))
	}
	return out, cur.Err()
}
func (r *MongoSessionRepo) Revoke(ctx context.Context, id string) error {
	_, err := r.sessions.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"revoked": true}})
	return err
}
func (r *MongoSessionRepo) RevokeAllByUser(ctx context.Context, uid valueobject.UserId) error {
	_, err := r.sessions.UpdateMany(ctx, bson.M{"user_id": uid.String()}, bson.M{"$set": bson.M{"revoked": true}})
	return err
}
func (r *MongoSessionRepo) Touch(ctx context.Context, id string) error {
	_, err := r.sessions.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"last_active_at": time.Now().UnixMilli()}})
	return err
}

// —— VerificationRepository ——

type verificationDoc struct {
	UserID         string `bson:"_id"`
	IDType         string `bson:"id_type"`
	IDDigest       string `bson:"id_digest"`
	Status         string `bson:"status"`
	ConsentAt      int64  `bson:"consent_at"`
	RetentionUntil int64  `bson:"retention_until"`
}

func toVerificationDoc(v entity.Verification) verificationDoc {
	return verificationDoc{
		UserID:         v.UserID().String(),
		IDType:         v.IDType(),
		IDDigest:       v.IDDigest(),
		Status:         string(v.Status()),
		ConsentAt:      v.ConsentAt().UnixMilli(),
		RetentionUntil: v.RetentionUntil().UnixMilli(),
	}
}

func fromVerificationDoc(d verificationDoc) entity.Verification {
	uid, _ := valueobject.NewUserId(d.UserID)
	v := entity.NewVerification(uid, d.IDType, d.IDDigest, time.UnixMilli(d.ConsentAt), time.UnixMilli(d.RetentionUntil))
	if d.Status == string(entity.VerificationRevoked) {
		v.Revoke()
	}
	return v
}

type MongoVerificationRepo struct{ verifs *mongo.Collection }

func NewMongoVerificationRepo(db *mongo.Database) *MongoVerificationRepo {
	return &MongoVerificationRepo{verifs: db.Collection("identity.verifications")}
}

func (r *MongoVerificationRepo) Save(ctx context.Context, v entity.Verification) (entity.Verification, error) {
	_, err := r.verifs.ReplaceOne(ctx, bson.M{"_id": v.UserID().String()}, toVerificationDoc(v), options.Replace().SetUpsert(true))
	return v, err
}
func (r *MongoVerificationRepo) FindByUser(ctx context.Context, uid valueobject.UserId) (entity.Verification, error) {
	var d verificationDoc
	if err := r.verifs.FindOne(ctx, bson.M{"_id": uid.String()}).Decode(&d); err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Verification{}, port.ErrNotFound
		}
		return entity.Verification{}, err
	}
	return fromVerificationDoc(d), nil
}
func (r *MongoVerificationRepo) Revoke(ctx context.Context, uid valueobject.UserId) error {
	_, err := r.verifs.UpdateOne(ctx, bson.M{"_id": uid.String()}, bson.M{
		"$set": bson.M{"status": "revoked", "id_digest": ""},
	})
	return err
}
func (r *MongoVerificationRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.verifs.DeleteMany(ctx, bson.M{"retention_until": bson.M{"$lt": before.UnixMilli()}})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

// —— NameClaimRepository ——

type nameClaimDoc struct {
	Name          string `bson:"_id"`
	OwnerUserID   string `bson:"owner_user_id"`
	ReleasedAt    *int64 `bson:"released_at,omitempty"`
	ReservedUntil *int64 `bson:"reserved_until,omitempty"`
}

type MongoNameClaimRepo struct{ claims *mongo.Collection }

func NewMongoNameClaimRepo(db *mongo.Database) *MongoNameClaimRepo {
	return &MongoNameClaimRepo{claims: db.Collection("identity.name_claims")}
}

func (r *MongoNameClaimRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.claims.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoNameClaimRepo) Claim(ctx context.Context, n valueobject.PublicName, uid valueobject.UserId) error {
	_, err := r.claims.InsertOne(ctx, nameClaimDoc{Name: n.String(), OwnerUserID: uid.String()})
	return err
}
func (r *MongoNameClaimRepo) Release(ctx context.Context, n valueobject.PublicName) error {
	now := time.Now()
	retention := time.Duration(config.Days("IDENTITY_NAME_RETENTION_DAYS", 90)) * 24 * time.Hour
	reserved := now.Add(retention)
	nowMs, reservedMs := now.UnixMilli(), reserved.UnixMilli()
	_, err := r.claims.UpdateOne(ctx, bson.M{"_id": n.String()}, bson.M{
		"$set": bson.M{"released_at": &nowMs, "reserved_until": &reservedMs},
	})
	return err
}
func (r *MongoNameClaimRepo) IsAvailable(ctx context.Context, n valueobject.PublicName) (bool, error) {
	var d nameClaimDoc
	if err := r.claims.FindOne(ctx, bson.M{"_id": n.String()}).Decode(&d); err != nil {
		if err == mongo.ErrNoDocuments {
			return true, nil // 无记录 = 可用
		}
		return false, err
	}
	if d.ReservedUntil != nil {
		return time.Now().UnixMilli() > *d.ReservedUntil, nil
	}
	return false, nil // 仍占用
}
