package database

import (
	"context"
	"fmt"

	"github.com/disgoorg/log"
	"github.com/ftqo/kirby/config"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func OpenDB(ctx context.Context, log log.Logger, c config.DBConfig) DB {
	log.Debug("opening database connection pool")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database)
	p, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Panic("failed to open connection pool: ", err)
	}
	return DB{pool: p}
}

func (db DB) Close(log log.Logger) {
	log.Info("gracefully closing database connection pool")
	db.pool.Close()
}

func (db DB) InitDatabase(ctx context.Context, log log.Logger) {
	log.Info("initializing database")
	db.createHstoreExtension(ctx, log)
	db.createGuildWelcomeTable(ctx, log)
	db.createKVTable(ctx, log)
}

func (db DB) createHstoreExtension(ctx context.Context, log log.Logger) {
	log.Debug("creating hstore extension if not exists")
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		log.Panic("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	CREATE EXTENSION IF NOT EXISTS hstore`
	_, err = conn.Exec(ctx, statement)
	if err != nil {
		log.Panicf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) createGuildWelcomeTable(ctx context.Context, log log.Logger) {
	log.Debug("creating guild_welcome table if not exists")
	conn, err := db.pool.Acquire(ctx)
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

func (db DB) createKVTable(ctx context.Context, log log.Logger) {
	log.Debug("creating kv_store table if not exists")
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		log.Panic("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	CREATE TABLE IF NOT EXISTS kv_store (
		name TEXT PRIMARY KEY,
		kv HSTORE
	)`
	_, err = conn.Exec(ctx, statement)
	if err != nil {
		log.Panicf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) UpsertKV(ctx context.Context, log log.Logger, name string, kv map[string]string) {
	log.Debug("inserting kv pair(s) for ", name)
	hstore := &pgtype.Hstore{
		Map:    make(map[string]pgtype.Text),
		Status: pgtype.Present,
	}
	for k := range kv {
		hstore.Map[k] = pgtype.Text{
			String: kv[k],
			Status: pgtype.Present,
		}
	}
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	INSERT INTO kv_store (name, kv)
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET kv = $2`
	_, err = conn.Exec(ctx, statement, name, hstore)
	if err != nil {
		log.Errorf("failed to execute %s: %v", statement, err)
	}
}

func (db DB) GetKV(ctx context.Context, log log.Logger, name string) (map[string]string, error) {
	log.Debug("getting kv pair(s) for ", name)
	hstore := pgtype.Hstore{}
	kv := make(map[string]string)
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
	}
	defer conn.Release()
	statement := `
	SELECT * FROM kv_store WHERE name = $1`
	row := conn.QueryRow(ctx, statement, name)
	err = row.Scan(nil, &hstore)
	if err != nil {
		return kv, err
	}
	if hstore.Status == pgtype.Null {
		return kv, err
	}
	for k := range hstore.Map {
		if hstore.Map[k].Status != pgtype.Present {
			return kv, err
		}
		kv[k] = hstore.Map[k].String
	}
	return kv, nil
}

func (db DB) InsertGuild(ctx context.Context, log log.Logger, guildID string) {
	log.Debug("inserting guild ", guildID)
	conn, err := db.pool.Acquire(ctx)
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

func (db DB) DeleteGuild(ctx context.Context, log log.Logger, guildID string) {
	log.Debug("deleting guild ", guildID)
	conn, err := db.pool.Acquire(ctx)
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

func (db DB) ResetGuild(ctx context.Context, log log.Logger, guildID string) {
	log.Debugf("resetting guild %s", guildID)
	db.DeleteGuild(ctx, log, guildID)
	db.InsertGuild(ctx, log, guildID)
}

func (db DB) GetGuildWelcome(ctx context.Context, log log.Logger, guildID string) (GuildWelcome, error) {
	log.Debug("getting guild welcome for ", guildID)
	gw := GuildWelcome{}
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection from pool: ", err)
		return gw, nil
	}
	defer conn.Release()
	statement := `
	SELECT * FROM guild_welcome WHERE guild_id = $1`
	row := conn.QueryRow(ctx, statement, guildID)
	err = row.Scan(&gw.GuildID, &gw.ChannelID, &gw.Type, &gw.Text, &gw.Image, &gw.ImageText)
	return gw, err
}

func (db DB) SetGuildWelcomeChannel(ctx context.Context, log log.Logger, guildID, channelID string) {
	log.Debug("setting guild welcome channel for ", guildID)
	conn, err := db.pool.Acquire(ctx)
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

func (db DB) SetGuildWelcomeType(ctx context.Context, log log.Logger, guildID, welcomeType string) {
	log.Debug("setting guild welcome type for ", guildID)
	conn, err := db.pool.Acquire(ctx)
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

func (db DB) SetGuildWelcomeText(ctx context.Context, log log.Logger, guildID, messageText string) {
	log.Debug("setting guild welcome text for ", guildID)
	conn, err := db.pool.Acquire(ctx)
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

func (db DB) SetGuildWelcomeImage(ctx context.Context, log log.Logger, guildID, image string) {
	log.Debug("setting guild welcome image for ", guildID)
	conn, err := db.pool.Acquire(ctx)
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

func (db DB) SetGuildWelcomeImageText(ctx context.Context, log log.Logger, guildID, imageText string) {
	log.Debug("setting guild welcome image text for ", guildID)
	conn, err := db.pool.Acquire(ctx)
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
