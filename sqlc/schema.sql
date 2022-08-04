CREATE TABLE kv_pairs
  (
     k VARCHAR PRIMARY KEY,
     v VARCHAR NOT NULL
  );

CREATE TABLE welcomes
  (
     guild_id       VARCHAR PRIMARY KEY,
     channel_id     VARCHAR NOT NULL,
     message_type   VARCHAR NOT NULL,
     message_text   VARCHAR NOT NULL,
     image_name     VARCHAR NOT NULL,
     image_title    VARCHAR NOT NULL,
     image_subtitle VARCHAR NOT NULL
  ); 