-- name: UpsertKV :exec
INSERT INTO kv_pairs (k, v)
	VALUES ($1, $2)
	ON CONFLICT (k) DO UPDATE
	SET v = $2;

-- name: GetV :one
SELECT v FROM kv_pairs WHERE k = $1;

-- name: InsertWelcome :exec
INSERT INTO welcomes (guild_id, channel_id, message_type, message_text, image_name, image_title, image_subtitle)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (guild_id) DO NOTHING;

-- name: GetWelcome :one
SELECT * FROM welcomes WHERE guild_id = $1;

-- name: SetWelcomeChannel :exec
UPDATE welcomes SET channel_id = $1 WHERE guild_id = $2;

-- name: SetWelcomeMessageType :exec
UPDATE welcomes SET message_type = $1 WHERE guild_id = $2;

-- name: SetWelcomeMessageText :exec
UPDATE welcomes SET message_text = $1 WHERE guild_id = $2;

-- name: SetWelcomeImageName :exec
UPDATE welcomes SET image_name = $1 WHERE guild_id = $2;

-- name: SetWelcomeImageTitle :exec
UPDATE welcomes SET image_title = $1 WHERE guild_id = $2;

-- name: SetWelcomeImageSubtitle :exec
UPDATE welcomes SET image_subtitle = $1 WHERE guild_id = $2;

-- name: DeleteWelcome :exec
DELETE FROM welcomes WHERE guild_id = $1;
