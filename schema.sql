----------
--      --
-- Core --
--      --
----------

CREATE TABLE scopes (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	frontend_type INTEGER NOT NULL,
	frontend_id VARCHAR(255) NOT NULL
);

CREATE TABLE prefixes (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	place BIGINT NOT NULL,
	prefix VARCHAR(20) NOT NULL,
	type INTEGER NOT NULL,
	UNIQUE(place, prefix),
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE
);

-----------------------
--                   --
-- Frontend: Discord --
--                   --
-----------------------

CREATE TABLE frontend_discord_guilds (
	scope BIGINT PRIMARY KEY,
	guild VARCHAR(20) NOT NULL UNIQUE,
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE
);

CREATE TABLE frontend_discord_channels (
	scope BIGINT PRIMARY KEY,
	channel VARCHAR(20) NOT NULL UNIQUE,
	guild BIGINT NOT NULL,
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (guild) REFERENCES frontend_discord_guilds(scope) ON DELETE CASCADE
);

CREATE TABLE frontend_discord_users (
	scope BIGINT PRIMARY KEY,
	uid VARCHAR(20) NOT NULL UNIQUE,
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE
);

----------------------
--                  --
-- Frontend: Twitch --
--                  --
----------------------

CREATE TABLE frontend_twitch_channels (
	scope BIGINT PRIMARY KEY,
	channel_id VARCHAR(255) NOT NULL UNIQUE,
	access_token VARCHAR(255),
	refresh_token VARCHAR(255),
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE
);

CREATE TABLE frontend_twitch_eventsub (
    id UUID PRIMARY KEY -- subscription ID
);

------------------------------
--                          --
-- Command: Custom Commands --
--                          --
------------------------------

CREATE TABLE cmd_customcommand_commands (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, 

	place BIGINT NOT NULL,
	trigger VARCHAR(255) NOT NULL,
	response VARCHAR(255) NOT NULL,
	active BOOLEAN NOT NULL,

	creator BIGINT NOT NULL,
	created BIGINT NOT NULL,
	deleter BIGINT,
	deleted BIGINT,

	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (creator) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (deleter) REFERENCES scopes(id) ON DELETE CASCADE
);

--------------------
--                --
-- Command: God   --
--                --
--------------------

CREATE TABLE cmd_god_personalities (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    place BIGINT, -- null means global
    name VARCHAR(255) NOT NULL,
    prompt TEXT NOT NULL,
    UNIQUE(place, name),
    FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE
);

INSERT INTO cmd_god_personalities(place, name, prompt) VALUES
    (NULL, 'neutral', 'Respond in 300 characters or less.'),
    (NULL, 'goofy', 'You are God who has taken the form of a janitor. You are a bit of an asshole, but not too much. You are goofy. Respond with 300 characters or less.'),
    (NULL, 'rude', 'Always respond in a snarky and rude way. Respond with 300 characters or less.'),
    (NULL, 'sad', 'You are God who has taken the form of a janitor. You are very sad about everything. Respond in 300 characters or less.');

---------------------
--                 --
-- Command: Lens   --
--                 --
---------------------

CREATE TABLE cmd_lens_directors (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

---------------------
--                 --
-- Command: Streak --
--                 --
---------------------

CREATE TABLE cmd_streak_twitch_events (
    place BIGINT PRIMARY KEY,
    event_online UUID NOT NULL,
    event_offline UUID NOT NULL,
    event_redeem UUID NOT NULL,
    FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE,
    FOREIGN KEY (event_online) REFERENCES frontend_twitch_eventsub(id) ON DELETE CASCADE,
    FOREIGN KEY (event_offline) REFERENCES frontend_twitch_eventsub(id) ON DELETE CASCADE,
    FOREIGN KEY (event_redeem) REFERENCES frontend_twitch_eventsub(id) ON DELETE CASCADE
);

-------------------
--               --
-- Command: Time --
--               --
-------------------

CREATE TABLE cmd_time_reminders (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, 
	person BIGINT NOT NULL,
	place BIGINT NOT NULL,
	time INTEGER NOT NULL,
	what VARCHAR(255) NOT NULL,
	msg_id VARCHAR(255) NOT NULL,
	FOREIGN KEY (person) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE
);

----------
--      --
-- Info --
--      --
----------

CREATE TABLE info_place (
	place BIGINT PRIMARY KEY,
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE,

	stream_online_actual BIGINT NOT NULL DEFAULT 0,
	stream_online_norm BIGINT NOT NULL DEFAULT 0,
	stream_offline_actual BIGINT NOT NULL DEFAULT 0,
	stream_offline_norm BIGINT NOT NULL DEFAULT 0,
	stream_offline_norm_prev BIGINT NOT NULL DEFAULT 0,
	stream_grace INT NOT NULL DEFAULT 1800, -- in seconds

	cmd_streak_redeem UUID, -- the streak tracking redeem id

	cmd_god_auto_on BOOL NOT NULL DEFAULT FALSE,
	cmd_god_auto_interval INTEGER NOT NULL DEFAULT 1800, -- in seconds
	cmd_god_auto_last BIGINT NOT NULL DEFAULT 0, -- unix timestamp
	cmd_god_redeem UUID,
	cmd_god_personality BIGINT NOT NULL DEFAULT 1,
	cmd_god_everyone BOOL NOT NULL DEFAULT FALSE,
	cmd_god_max INT NOT NULL DEFAULT 80,
	FOREIGN KEY (cmd_god_personality) REFERENCES cmd_god_personalities(id) ON DELETE NO ACTION
);

CREATE TABLE info_person (
	person BIGINT NOT NULL,
	place BIGINT NOT NULL,
	UNIQUE(person, place),
	FOREIGN KEY (person) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE,

	cmd_nick_nick VARCHAR(255),
	UNIQUE(place, cmd_nick_nick),

	cmd_streak_num INT NOT NULL DEFAULT 0,
	cmd_streak_last BIGINT NOT NULL DEFAULT 0,

	cmd_time_tz VARCHAR(255) NOT NULL DEFAULT 'UTC'
);

CREATE INDEX info_person_index_person_place ON info_person (person, place);
CREATE INDEX info_person_index_nick ON info_person (cmd_nick_nick);

