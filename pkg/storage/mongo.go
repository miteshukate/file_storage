package storage

import (
	"context"
	"errors"
	"log"
	"time"

	"file_storage/pkg/api"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoClient creates a MongoDB client using the provided URI.
func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return client, nil
}

// MongoContentRepository implements api.ContentRepository against MongoDB.
type MongoContentRepository struct {
	client     *mongo.Client
	ctx        context.Context
	database   string
	collection string
}

func (r *MongoContentRepository) FindByFilter(filter interface{}) ([]api.Content, error) {
	cur, err := r.coll().Find(r.ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(r.ctx)
	var out []api.Content
	for cur.Next(r.ctx) {
		var c api.Content
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, cur.Err()
}

func NewMongoContentRepository() *MongoContentRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := NewMongoClient(ctx, "mongodb://localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	// Use Background for repository operations; apply per-call timeouts as needed.
	return &MongoContentRepository{client: client, database: "local", collection: "contents", ctx: context.Background()}
}

func (r *MongoContentRepository) coll() *mongo.Collection {
	return r.client.Database(r.database).Collection(r.collection)
}

// Create inserts a new content document and returns the persisted entity with its ID.
func (r *MongoContentRepository) Create(content *api.Content) (*api.Content, error) {
	// ensure timestamps
	if content.CreatedAt.IsZero() {
		content.CreatedAt = time.Now()
	}
	content.UpdatedAt = time.Now()
	res, err := r.coll().InsertOne(r.ctx, content)
	if err != nil {
		return nil, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		content.ID = oid
	}
	return content, nil
}

// Update replaces fields on a content document and returns the updated entity.
func (r *MongoContentRepository) Update(id primitive.ObjectID, filter api.Filter, content *api.Content) (*api.Content, error) {
	content.UpdatedAt = time.Now()
	baseFilter := bson.M{"_id": id}
	// Build update doc without _id and without CreatedAt modification
	set := bson.M{
		"parentId":     content.ParentID,
		"name":         content.Name,
		"type":         content.Type,
		"size":         content.Size,
		"contentType":  content.ContentType,
		"status":       content.Status,
		"etag":         content.ETag,
		"lastModified": content.LastModified,
		"updatedAt":    content.UpdatedAt,
	}
	log.Println(baseFilter, set)
	update := bson.M{"$set": content}
	_, err := r.coll().UpdateOne(r.ctx, baseFilter, update)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// FindAll returns all contents. In a real app, add pagination.
func (r *MongoContentRepository) FindAll() ([]api.Content, error) {
	cur, err := r.coll().Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			log.Println(err)
		}
	}(cur, r.ctx)
	var out []api.Content
	for cur.Next(r.ctx) {
		var c api.Content
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, cur.Err()
}

// FindById finds a content by its UUID (legacy compatibility). New records use ObjectID.
func (r *MongoContentRepository) FindById(id primitive.ObjectID) (*api.Content, error) {
	var c api.Content
	// Find by _id as ObjectId
	err := r.coll().FindOne(r.ctx, bson.M{"_id": id}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindByParentId finds content by parent ID. In a real app, you may want to return multiple.
func (r *MongoContentRepository) FindByParentId(parentId string) (*api.Content, error) {
	var c api.Content
	err := r.coll().FindOne(r.ctx, bson.M{"parentId": parentId}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindByIds finds multiple contents by their ObjectIDs
func (r *MongoContentRepository) FindByIds(ids []string) ([]api.Content, error) {
	if len(ids) == 0 {
		return []api.Content{}, nil
	}

	// Convert string IDs to ObjectIDs
	objectIds := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			log.Printf("invalid ObjectID format: %s", id)
			continue
		}
		objectIds[i] = oid
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	cur, err := r.coll().Find(r.ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(r.ctx)

	var out []api.Content
	for cur.Next(r.ctx) {
		var c api.Content
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, cur.Err()
}

// EnsureIndexes creates helpful indexes for performance.
func (r *MongoContentRepository) EnsureIndexes(ctx context.Context) error {
	mdl := r.coll()
	// index on parentId for quick lookup
	_, err := mdl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "parentId", Value: 1}},
		Options: options.Index().SetBackground(true),
	})
	return err
}

// WithTimeout wraps a context with a default timeout for Mongo operations.
func WithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		d = 5 * time.Second
	}
	return context.WithTimeout(ctx, d)
}
