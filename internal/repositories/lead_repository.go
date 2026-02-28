package repositories

import (
	"context"
	"fmt"
	"time"

	"leadflowcrm/internal/config"
	"leadflowcrm/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository() *UserRepository {
	db := config.GetDatabase()
	return &UserRepository{
		collection: db.Database.Collection("users"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user})
	return err
}

func (r *UserRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

type LeadRepository struct {
	collection *mongo.Collection
}

func NewLeadRepository() *LeadRepository {
	db := config.GetDatabase()
	return &LeadRepository{
		collection: db.Database.Collection("leads"),
	}
}

func (r *LeadRepository) Create(ctx context.Context, lead *models.Lead) error {
	lead.CreatedAt = time.Now()
	lead.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, lead)
	if err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}
	lead.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *LeadRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Lead, error) {
	var lead models.Lead
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "deleted": false}).Decode(&lead)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) Find(ctx context.Context, filter models.LeadFilter) ([]models.Lead, int64, error) {
	query := bson.M{"deleted": false}

	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.AssignedTo != "" {
		oid, err := primitive.ObjectIDFromHex(filter.AssignedTo)
		if err == nil {
			query["assigned_to"] = oid
		}
	}
	if filter.Source != "" {
		query["source"] = filter.Source
	}
	if filter.Search != "" {
		query["$or"] = []bson.M{
			{"name": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"company": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var leads []models.Lead
	if err := cursor.All(ctx, &leads); err != nil {
		return nil, 0, err
	}

	return leads, total, nil
}

func (r *LeadRepository) Update(ctx context.Context, lead *models.Lead) error {
	lead.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": lead.ID}, bson.M{"$set": lead})
	return err
}

func (r *LeadRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"deleted":    true,
			"deleted_at": now,
			"updated_at": now,
		},
	})
	return err
}

func (r *LeadRepository) GetStats(ctx context.Context) (*models.LeadStats, error) {
	stats := &models.LeadStats{
		ByStatus: make(map[string]int),
		BySource: make(map[string]int),
	}

	total, err := r.collection.CountDocuments(ctx, bson.M{"deleted": false})
	if err != nil {
		return nil, err
	}
	stats.Total = int(total)

	pipeline := []bson.M{
		{"$match": bson.M{"deleted": false}},
		{"$group": bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats.ByStatus[result.ID] = result.Count
	}

	pipeline = []bson.M{
		{"$match": bson.M{"deleted": false}},
		{"$group": bson.M{"_id": "$source", "count": bson.M{"$sum": 1}}},
	}

	cursor, err = r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats.BySource[result.ID] = result.Count
	}

	assigned, _ := r.collection.CountDocuments(ctx, bson.M{"deleted": false, "assigned_to": bson.M{"$ne": nil}})
	stats.Assigned = int(assigned)
	stats.Unassigned = stats.Total - stats.Assigned

	valuePipeline := []bson.M{
		{"$match": bson.M{"deleted": false}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$value"}}},
	}

	cursor, err = r.collection.Aggregate(ctx, valuePipeline)
	if err == nil && cursor.Next(ctx) {
		var result struct {
			Total float64 `bson:"total"`
		}
		cursor.Decode(&result)
		stats.TotalValue = result.Total
	}

	return stats, nil
}

func (r *LeadRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "email", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "phone", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "assigned_to", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

type LeadHistoryRepository struct {
	collection *mongo.Collection
}

func NewLeadHistoryRepository() *LeadHistoryRepository {
	db := config.GetDatabase()
	return &LeadHistoryRepository{
		collection: db.Database.Collection("lead_history"),
	}
}

func (r *LeadHistoryRepository) Create(ctx context.Context, history *models.LeadHistory) error {
	history.CreatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to create lead history: %w", err)
	}
	history.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *LeadHistoryRepository) FindByLeadID(ctx context.Context, leadID primitive.ObjectID) ([]models.LeadHistory, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"lead_id": leadID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var histories []models.LeadHistory
	if err := cursor.All(ctx, &histories); err != nil {
		return nil, err
	}
	return histories, nil
}

type LeadNoteRepository struct {
	collection *mongo.Collection
}

func NewLeadNoteRepository() *LeadNoteRepository {
	db := config.GetDatabase()
	return &LeadNoteRepository{
		collection: db.Database.Collection("lead_notes"),
	}
}

func (r *LeadNoteRepository) Create(ctx context.Context, note *models.LeadNote) error {
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	note.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *LeadNoteRepository) FindByLeadID(ctx context.Context, leadID primitive.ObjectID) ([]models.LeadNote, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"lead_id": leadID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notes []models.LeadNote
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *LeadNoteRepository) Update(ctx context.Context, note *models.LeadNote) error {
	note.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": note.ID}, bson.M{"$set": note})
	return err
}

func (r *LeadNoteRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
