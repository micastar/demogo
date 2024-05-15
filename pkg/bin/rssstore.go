package bin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/micastar/discord-feed/pkg/global"
	"github.com/micastar/discord-feed/pkg/util"
	"github.com/mmcdole/gofeed"
)

type FeedStore struct {
	conn *pgx.Conn
}

func NewFeedStore(connString string) (*FeedStore, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return &FeedStore{conn: conn}, nil
}

func (fs *FeedStore) Close() {
	fs.conn.Close(context.Background())
}

func (fs *FeedStore) SaveItem(item *gofeed.Item) error {

	query := `
		INSERT INTO contents (g_id, title, descrp, logdate, link, cates)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON
			CONFLICT (g_id,logdate)
		DO NOTHING
	`

	published := convertTime(item.Published)

	_, err := fs.conn.Exec(context.Background(), query, item.GUID, item.Title, item.Description, published, item.Link, item.Categories)
	if err != nil {
		return fmt.Errorf("unable to save item: %v", err)
	}
	return nil
}

func (fs *FeedStore) ReadDataGId() (*string, error) {

	var id string

	query := `
			select g_id from contents order by g_id desc limit 1
	`

	row := fs.conn.QueryRow(context.Background(), query)
	if err := row.Scan(&id); err != nil {
		return nil, errors.New("Can't get the id from db")
	}

	return &id, nil

}

type Post struct {
	Title   string
	Descrp  string
	Logdate string
	Link    string
	Cates   []string
}

func (fs *FeedStore) ReadDataWithLimit() ([]*Post, error) {
	var res []*Post
	var p Post

	query := `
		select title, descrp, logdate, link, cates from contents order by g_id desc limit $1
	`
	rows, _ := fs.conn.Query(context.Background(), query, global.Limit)

	_, err := pgx.ForEachRow(rows, []any{&p.Title, &p.Descrp, &p.Logdate, &p.Link, &p.Cates}, func() error {
		res = append(res, &Post{
			Title:   p.Title,
			Descrp:  p.Descrp,
			Logdate: p.Logdate,
			Link:    p.Link,
			Cates:   p.Cates,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error Read Data Interval Time: %v", err)
	}

	util.ReverseItem(res)

	return res, nil
}

func (fs *FeedStore) ReadLatestData(conId *string) ([]*Post, error) {

	var res []*Post
	var p Post
	var id string

	query := `
		select g_id, title, descrp, logdate, link, cates from contents where g_id > $1 order by g_id limit $2
	`

	rows, _ := fs.conn.Query(context.Background(), query, *conId, global.Limit)

	_, err := pgx.ForEachRow(rows, []any{&id, &p.Title, &p.Descrp, &p.Logdate, &p.Link, &p.Cates}, func() error {
		res = append(res, &Post{
			Title:   p.Title,
			Descrp:  p.Descrp,
			Logdate: p.Logdate,
			Link:    p.Link,
			Cates:   p.Cates,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Error Read Latest Data: %v", err)
	}

	if id == "" {
		log.Printf("Read Latest Data Id is Empty: %v", id)
		return res, nil
	}

	*conId = id

	return res, nil
}

func convertTime(localTime string) string {
	midTime, _ := time.Parse(time.RFC1123, localTime)
	return midTime.Format(global.DefaultTimeFormat)
}
