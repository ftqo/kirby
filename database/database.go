package database

import (
	"context"
	"fmt"

	"github.com/ftqo/kirby/config"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type DB struct {
	*pgxpool.Pool
}

func OpenDB(ctx context.Context, log *logrus.Logger, c config.DBConfig) DB {
	log.Info("opening database connection pool")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database)
	p, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Panic("failed to open connection pool: ", err)
	}
	return DB{Pool: p}
}

func (db DB) CloseDB(log *logrus.Logger) {
	log.Info("closing database connection pool")
	db.Close()
}

func (db DB) InitDatabase(ctx context.Context, log *logrus.Logger) {
	log.Info("initializing database")
	db.createGuildWelcomeTable(ctx, log)
	db.createKVTable(ctx, log)
}

func (db DB) createGuildWelcomeTable(ctx context.Context, log *logrus.Logger) {
	log.Info("creating guild_welcome table if not exists")
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Panic("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	CREATE TABLE IF NOT EXISTS guild_welcome (
		guild_id TEXT PRIMARY KEY,
		channel_id TEXT NOT NULL,
		type TEXT NOT NULL,
		message_text TEXT NOT NULL,
		image TEXT NOT NULL,
		image_text TEXT NOT NULL
	)`
	_, err = conn.Exec(ctx, statement)
	if err != nil {
		log.Panicf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) createKVTable(ctx context.Context, log *logrus.Logger) {
	log.Info("creating key_value table if not exists")
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Panic("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	CREATE TABLE IF NOT EXISTS key_value (
		name TEXT PRIMARY KEY,
		kv HSTORE
	)`
	_, err = conn.Exec(ctx, statement)
	if err != nil {
		log.Panicf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) InsertGuild(ctx context.Context, log *logrus.Logger, guildID string) {
	log.Infof("inserting guild %s", guildID)
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	dgw := NewDefaultGuildWelcome()
	statement := `
	INSERT INTO guild_welcome (guild_id, channel_id, type, message_text, image, image_text)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (guild_id) DO NOTHING`
	_, err = conn.Exec(ctx, statement, guildID, dgw.ChannelID, dgw.Type, dgw.Text, dgw.Image, dgw.ImageText)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) DeleteGuild(ctx context.Context, log *logrus.Logger, guildID string) {
	log.Infof("deleting guild %s", guildID)
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	DELETE FROM guild_welcome WHERE guild_id = $1`
	_, err = conn.Exec(ctx, statement, guildID)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) ResetGuild(ctx context.Context, log *logrus.Logger, guildID string) {
	log.Infof("resetting guild %s", guildID)
	db.DeleteGuild(ctx, log, guildID)
	db.InsertGuild(ctx, log, guildID)
}

func (db DB) GetGuildWelcome(ctx context.Context, log *logrus.Logger, guildID string) (GuildWelcome, error) {
	log.Infof("getting guild welcome for %s", guildID)
	gw := GuildWelcome{}
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
		return gw, nil
	}
	defer conn.Release()
	statement := `
	SELECT * FROM guild_welcome WHERE guild_id = $1`
	row := conn.QueryRow(ctx, statement, guildID)
	err = row.Scan(&gw.GuildID, &gw.ChannelID, &gw.Type, &gw.Text, &gw.Image, &gw.ImageText)
	if err != nil {
		log.Errorf("failed to query %s: %v", statement, err)
	}
	return gw, err
}

func (db DB) SetGuildWelcomeChannel(ctx context.Context, log *logrus.Logger, guildID, channelID string) {
	log.Infof("setting guild welcome channel for %s", guildID)
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	UPDATE guild_welcome SET channel_id = $1 WHERE guild_id = $2`
	_, err = conn.Exec(ctx, statement, channelID, guildID)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) SetGuildWelcomeType(ctx context.Context, log *logrus.Logger, guildID, welcomeType string) {
	log.Infof("setting guild welcome type for %s", guildID)
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	UPDATE guild_welcome SET type = $1 WHERE guild_id = $2`
	_, err = conn.Exec(ctx, statement, welcomeType, guildID)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) SetGuildWelcomeText(ctx context.Context, log *logrus.Logger, guildID, messageText string) {
	log.Infof("setting guild welcome text for %s", guildID)
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	UPDATE guild_welcome SET message_text = $1 WHERE guild_id = $2`
	_, err = conn.Exec(ctx, statement, messageText, guildID)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) SetGuildWelcomeImage(ctx context.Context, log *logrus.Logger, guildID, image string) {
	log.Infof("setting guild welcome image for %s", guildID)
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	UPDATE guild_welcome SET image = $1 WHERE guild_id = $2`
	_, err = conn.Exec(ctx, statement, image, guildID)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) SetGuildWelcomeImageText(ctx context.Context, log *logrus.Logger, guildID, imageText string) {
	log.Infof("setting guild welcome image text for %s", guildID)
	conn, err := db.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	UPDATE guild_welcome SET image_text = $1 WHERE guild_id = $2`
	_, err = conn.Exec(ctx, statement, imageText, guildID)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) UpsertSession(ctx context.Context, log *logrus.Logger, s Session) {

}