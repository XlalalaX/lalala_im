package AddReq

import (
	"context"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	db "lalala_im/data"
	MongoModel2 "lalala_im/data/model/MongoModel"
)

type IMongoAddReqRepo interface {
	Create(ctx context.Context, addReq *MongoModel2.AddReq) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, addReq *MongoModel2.AddReq) error
	GetListByRecvID(ctx context.Context, recvId string) ([]*MongoModel2.AddReq, error)
	GetListBySendID(ctx context.Context, sendId string) ([]*MongoModel2.AddReq, error)
	GetListByGroupID(ctx context.Context, groupId string) ([]*MongoModel2.AddReq, error)
	GetListByGroupIDAndSendID(ctx context.Context, groupId string, sendId string) ([]*MongoModel2.AddReq, error)
	GetListByRecvIDAndSendID(ctx context.Context, recvId string, sendId string) ([]*MongoModel2.AddReq, error)
	GetInfoByID(ctx context.Context, id string) (*MongoModel2.AddReq, error)
}

type addReqRepo struct {
	db.IMdBaseRepo
}

func (a addReqRepo) GetInfoByID(ctx context.Context, id string) (*MongoModel2.AddReq, error) {
	out := &MongoModel2.AddReq{}
	err := a.FindOne(ctx, id, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (a addReqRepo) GetListByRecvIDAndSendID(ctx context.Context, recvId string, sendId string) ([]*MongoModel2.AddReq, error) {
	results := []*MongoModel2.AddReq{}
	filter := bson.M{"recv_id": recvId, "send_id": sendId}
	err := a.QueryMany(ctx, &results, filter, bson.M{})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (a addReqRepo) Create(ctx context.Context, addReq *MongoModel2.AddReq) error {
	return a.Save(ctx, addReq)
}

func (a addReqRepo) Update(ctx context.Context, id string, addReq *MongoModel2.AddReq) error {
	return a.UpdateById(ctx, id, addReq)
}

func (a addReqRepo) GetListByRecvID(ctx context.Context, recvId string) ([]*MongoModel2.AddReq, error) {
	results := []*MongoModel2.AddReq{}
	filter := bson.M{"recv_id": recvId}
	err := a.QueryMany(ctx, &results, filter, bson.M{})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (a addReqRepo) GetListBySendID(ctx context.Context, sendId string) ([]*MongoModel2.AddReq, error) {
	results := []*MongoModel2.AddReq{}
	filter := bson.M{"send_id": sendId}
	err := a.QueryMany(ctx, &results, filter, bson.M{})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (a addReqRepo) GetListByGroupID(ctx context.Context, groupId string) ([]*MongoModel2.AddReq, error) {
	results := []*MongoModel2.AddReq{}
	filter := bson.M{"group_id": groupId}
	err := a.QueryMany(ctx, &results, filter, bson.M{})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (a addReqRepo) GetListByGroupIDAndSendID(ctx context.Context, groupId string, sendId string) ([]*MongoModel2.AddReq, error) {
	results := []*MongoModel2.AddReq{}
	filter := bson.M{"group_id": groupId, "send_id": sendId}
	err := a.QueryMany(ctx, &results, filter, bson.M{})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func NewIMongoAddReqRepo(database *qmgo.Database) IMongoAddReqRepo {
	addReq := MongoModel2.AddReq{}
	IRepo := &addReqRepo{&db.MdBaseRepo{
		Db:         database,
		Collection: addReq.TableName(),
	}}

	return IRepo
}
