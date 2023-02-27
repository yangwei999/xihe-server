package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewCollection(name string) collection {
	return collection{name}
}

type collection struct {
	name string
}

func (c collection) Collection() *mongo.Collection {
	return cli.collection(c.name)
}

func (c collection) IsDocNotExists(err error) bool {
	return isDocNotExists(err)
}

func (c collection) IsDocExists(err error) bool {
	return isDocExists(err)
}

func (c collection) ObjectIdFilter(s string) (bson.M, error) {
	return objectIdFilter(s)
}

func (c collection) AppendElemMatchToFilter(array string, exists bool, cond, filter bson.M) {
	appendElemMatchToFilter(array, exists, cond, filter)
}

func (c collection) GetDoc(
	ctx context.Context, filterOfDoc, project bson.M,
	result interface{},
) error {
	return cli.getDoc(ctx, c.name, filterOfDoc, project, result)
}

func (c collection) GetDocs(
	ctx context.Context, filterOfDoc, project bson.M,
	result interface{},
) error {
	return cli.getDocs(ctx, c.name, filterOfDoc, project, result)
}

func (c collection) DeleteDoc(
	ctx context.Context, filterOfDoc bson.M,
) error {
	return cli.deleteDoc(ctx, c.name, filterOfDoc)
}

func (c collection) NewDocIfNotExist(
	ctx context.Context, filterOfDoc, docInfo bson.M,
) (string, error) {
	return cli.newDocIfNotExist(ctx, c.name, filterOfDoc, docInfo)
}

func (c collection) PushArrayElem(
	ctx context.Context, array string,
	filterOfDoc, value bson.M,
) error {
	return cli.pushArrayElem(ctx, c.name, array, filterOfDoc, value)
}

func (c collection) UpdateArrayElem(
	ctx context.Context, array string,
	filterOfDoc, filterOfArray, updateCmd bson.M,
	version int, t int64,
) (bool, error) {
	return cli.updateArrayElem(
		ctx, c.name, array, filterOfDoc, filterOfArray, updateCmd,
		version, t,
	)
}

func (c collection) UpdateDoc(
	ctx context.Context, filterOfDoc, update bson.M, op string, version int,
) error {
	return cli.updateDoc(ctx, c.name, filterOfDoc, update, op, version)
}

func (c collection) ModifyArrayElem(
	ctx context.Context, array string,
	filterOfDoc, filterOfArray, updateCmd bson.M, op string,
) (bool, error) {
	return cli.modifyArrayElemWithoutVersion(ctx, c.name, array,
		filterOfDoc, filterOfArray, updateCmd, op,
	)
}
