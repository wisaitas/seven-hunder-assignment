// pkg/httpx/pagination_mongo.go
package httpx

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ProbeNextPagesMongo probes ahead to see how many pages are available
// Similar to ProbeNextPages but works with MongoDB
func ProbeNextPagesMongo(
	ctx context.Context,
	collection *mongo.Collection,
	filter bson.M,
	sortField string,
	sortOrder int, // 1 for ascending, -1 for descending
	page, pageSize, wantPages int,
) (int, error) {
	if wantPages <= 0 {
		return 0, nil
	}

	// How many records do we need to check ahead
	need := wantPages*pageSize + 1

	// Create projection to only fetch _id (minimal data transfer)
	projection := bson.M{"_id": 1}

	// Create find options with sort, skip, and limit
	findOptions := options.Find().
		SetProjection(projection).
		SetSort(bson.M{sortField: sortOrder}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(need))

	// Execute the query
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	// Count results
	var count int
	for cursor.Next(ctx) {
		count++
	}

	if err := cursor.Err(); err != nil {
		return 0, err
	}

	// Calculate remaining records after current page
	remain := count
	if remain > pageSize {
		remain = remain - pageSize
	} else {
		// No data beyond current page
		return 0, nil
	}

	// Calculate available pages
	availPages := remain / pageSize
	if availPages > wantPages {
		availPages = wantPages
	}
	if availPages < 0 {
		availPages = 0
	}

	return availPages, nil
}
