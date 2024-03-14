package storage

import (
	"context"
	"encoding/json"
	"highload-arch/pkg/common"
	"log"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/exp/slices"
)

type PostRequest struct {
	ID           string    `pg:"id"`
	AuthorUserID string    `pg:"author_user_id"`
	CreatedAt    time.Time `pg:"created_at"`
	UpdatedAt    time.Time `pg:"updated_at"`
	Text         string    `pg:"text"`
	//UserID       string
}

const CACHE_TTL = 10 // Cache TTl in seconds

func (req *PostRequest) dbAddPost(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO posts (author_user_id, text, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		req.AuthorUserID, req.Text, req.CreatedAt, req.UpdatedAt)

	return err
}

func (req *PostRequest) dbDeletePost(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`DELETE FROM posts WHERE id = $1`, req.ID)

	return err
}

func (req *PostRequest) dbUpdatePost(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx,
		`UPDATE posts SET text = $1, updated_at = $2 WHERE id = $3`, req.Text, time.Now(), req.ID)

	return err
}

func dbFeedPosts(ctx context.Context, userID string, offset, limit int) ([]PostRequest, error) {
	res := []PostRequest{}

	rows, err := db.Query(ctx,
		`SELECT id, author_user_id, created_at, updated_at, text FROM posts WHERE author_user_id in (SELECT friend_id FROM friends WHERE id = $1)`, userID)

	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	if offset >= len(res) {
		return nil, nil
	}
	if offset+limit > len(res) {
		return res[offset:], nil
	}
	return res[offset:limit], nil
}

/* Load last 1000 updated posts from DB */
func dbLoadPosts(ctx context.Context, limit int) ([]PostRequest, error) {
	res := []PostRequest{}

	rows, err := db.Query(ctx,
		`SELECT id, author_user_id, created_at, updated_at, text FROM posts WHERE author_user_id in (SELECT friend_id FROM friends) ORDER BY updated_at LIMIT $1`, limit)

	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	return res, err
}

func dbGetPost(ctx context.Context, id string) (*PostRequest, error) {
	res := []PostRequest{}

	rows, err := db.Query(ctx,
		`SELECT id, author_user_id, created_at, updated_at, text from posts WHERE id = $1`, id)

	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err = pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, common.ErrPostNotFound
	}

	return &res[0], err
}

func CreatePost(ctx context.Context, userID string, text string) error {
	req := &PostRequest{AuthorUserID: userID, Text: text, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := req.dbAddPost(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	err = QueuePostCreatedMessage(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func QueuePostCreatedMessage(ctx context.Context, req *PostRequest) error {
	channel, err := rbmq.Channel()
	if err != nil {
		log.Println("RBMQ: Channel creation failed")
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"createdPosts", // name
		"topic",        // type
		false,          // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)

	if err != nil {
		log.Println("Cannot create exchange")
		return err
	}
	reqBytes, err := json.Marshal(*req)
	if err != nil {
		log.Println("Cannot marshal post request to bytes array")
		return err
	}
	// rounting key: userID.postID
	err = channel.PublishWithContext(ctx,
		"createdPosts",              // exchange
		req.AuthorUserID+"."+req.ID, // routing key
		false,                       // mandatory
		false,                       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        reqBytes,
		})

	return nil
}

func DeletePost(ctx context.Context, id string) error {
	req := &PostRequest{ID: id}
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := req.dbDeletePost(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func GetPost(ctx context.Context, id string) (*PostRequest, error) {
	post, err := dbGetPost(ctx, id)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func UpdatePost(ctx context.Context, id string, text string) error {
	req := &PostRequest{ID: id, Text: text}
	_, err := HandleInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) (interface{}, error) {
		err := req.dbUpdatePost(ctx, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func FeedPosts(ctx context.Context, userID string, offset, limit int) ([]PostRequest, error) {
	friends, err := cacheGetFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	var posts []PostRequest
	if len(friends) != 0 {
		posts, err = cacheGetPosts(ctx, friends, offset, limit)
		if err != nil {
			return nil, err
		}
	}
	if len(posts) == 0 {
		posts, err = dbFeedPosts(ctx, userID, offset, limit)
		if err != nil {
			return nil, err
		}
	}
	return posts, nil
}

func cacheGetFriends(ctx context.Context, userID string) ([]string, error) {
	var friends []string

	iter := cache.Scan(ctx, 0, "user_friends:*", 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		values, err := cache.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		if values["user_id"] != userID {
			continue
		}

		friend := values["friend_id"]
		friends = append(friends, friend)

	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return friends, nil
}

func cacheGetPosts(ctx context.Context, friends []string, offset, limit int) ([]PostRequest, error) {
	var posts []PostRequest
	iter := cache.Scan(ctx, 0, "post:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		values, err := cache.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		if !slices.Contains(friends, values["author_user_id"]) {
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, values["created_at"])
		if err != nil {
			return nil, err
		}
		updatedAt, err := time.Parse(time.RFC3339, values["updated_at"])
		if err != nil {
			return nil, err
		}
		post := PostRequest{ID: values["post_id"],
			AuthorUserID: values["author_user_id"],
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			Text:         values["text"],
		}

		posts = append(posts, post)

	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	if offset >= len(posts) {
		return nil, nil
	}
	if offset+limit > len(posts) {
		return posts[offset:], nil
	}
	return posts[offset:limit], nil
}

func cacheUpdatePosts(ctx context.Context) {
	posts, err := dbLoadPosts(ctx, 1000)
	if err != nil {
		log.Println("Load posts failed: ", err)
	}

	for _, post := range posts {
		postSettings := map[string]string{"post_id": post.ID, "author_user_id": post.AuthorUserID, "created_at": post.CreatedAt.Format(time.RFC3339), "updated_at": post.UpdatedAt.Format(time.RFC3339), "text": post.Text}
		for k, v := range postSettings {
			err := cache.HSet(ctx, "post:"+post.ID, k, v).Err()
			if err != nil {
				log.Println("Cache update failed: ", err)
				return
			}
		}
	}

	friends, err := GetFriends(ctx)
	if err == nil {
		for _, friend := range friends {
			friendSettings := map[string]string{"user_id": friend.ID, "friend_id": friend.FriendID}
			for k, v := range friendSettings {
				err := cache.HSet(ctx, "user_friends:"+friend.ID, k, v).Err()
				if err != nil {
					log.Println("Cache update failed: ", err)
					return
				}
			}
		}
	}
}

func CacheUpdatePosts(ctx context.Context) {
	ticker := time.NewTicker(CACHE_TTL * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				cacheUpdatePosts(ctx)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
