----------
--      --
-- Core --
--      --
----------

CREATE TABLE IF NOT EXISTS scopes (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	frontend_type INTEGER NOT NULL,
	frontend_id VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS prefixes (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	place BIGINT NOT NULL,
	prefix VARCHAR(20) NOT NULL,
	type INTEGER NOT NULL,
	UNIQUE(place, prefix),
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS settings_place (
	place BIGINT PRIMARY KEY,
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE,

	cmd_tts_subonly BOOLEAN NOT NULL DEFAULT FALSE,

	cmd_god_reply_on BOOL NOT NULL DEFAULT FALSE,
	cmd_god_reply_interval INTEGER NOT NULL DEFAULT 1800, -- in seconds
	cmd_god_reply_last INTEGER NOT NULL DEFAULT 0 -- unix timestamp of last reply
);

CREATE TABLE IF NOT EXISTS settings_person (
	person BIGINT NOT NULL,
	place BIGINT NOT NULL,
	UNIQUE(person, place),
	FOREIGN KEY (person) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE,

	cmd_nick_nick VARCHAR(255),
	UNIQUE(place, cmd_nick_nick),

	cmd_time_tz VARCHAR(255) NOT NULL DEFAULT 'UTC',

	cmd_tts_voice VARCHAR(255) NOT NULL DEFAULT (ARRAY[
		-- DISNEY VOICES
		'en_us_ghostface',       -- Ghost Face
		'en_us_chewbacca',       -- Chewbacca
		'en_us_c3po',            -- C3PO
		'en_us_stitch',          -- Stitch
		'en_us_stormtrooper',    -- Stormtrooper
		'en_us_rocket',          -- Rocket
		'en_female_madam_leota', -- Madame Leota
		'en_male_ghosthost',     -- Ghost Host
		'en_male_pirate',        -- Pirate

		-- ENGLISH VOICES
		'en_au_001', -- English AU - Female
		'en_au_002', -- English AU - Male
		'en_uk_001', -- English UK - Male 1
		'en_uk_003', -- English UK - Male 2
		'en_us_001', -- English US - Female 1
		'en_us_002', -- English US - Female 2
		'en_us_006', -- English US - Male 1
		'en_us_007', -- English US - Male 2
		'en_us_009', -- English US - Male 3
		'en_us_010', -- English US - Male 4

		-- EUROPE VOICES
		'fr_001', -- French - Male 1
		'fr_002', -- French - Male 2
		'de_001', -- German - Female
		'de_002', -- German - Male
		'es_002', -- Spanish - Male

		-- AMERICA VOICES
		'es_mx_002', -- Spanish MX - Male
		'br_001',    -- Portuguese BR - Female 1
		'br_003',    -- Portuguese BR - Female 2
		'br_004',    -- Portuguese BR - Female 3
		'br_005',    -- Portuguese BR - Male

		-- ASIA VOICES
		'id_001', -- Indonesian - Female
		'jp_001', -- Japanese - Female 1
		'jp_003', -- Japanese - Female 2
		'jp_005', -- Japanese - Female 3
		'jp_006', -- Japanese - Male
		'kr_002', -- Korean - Male 1
		'kr_003', -- Korean - Female
		'kr_004', -- Korean - Male 2

		-- OTHER
		'en_male_narration',   -- Narrator
		'en_male_funny',       -- Wacky
		'en_female_emotional', -- Peaceful
		'en_male_cody'         -- Serious
	])[floor(random() * 41 + 1)]
);

CREATE INDEX IF NOT EXISTS settings_person_index_person_place ON settings_person (person, place);
CREATE INDEX IF NOT EXISTS settings_person_index_nick ON settings_person (cmd_nick_nick);

-----------------------
--                   --
-- Frontend: Discord --
--                   --
-----------------------

CREATE TABLE IF NOT EXISTS frontend_discord_guilds (
	scope BIGINT PRIMARY KEY,
	guild VARCHAR(20) NOT NULL UNIQUE,
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS frontend_discord_channels (
	scope BIGINT PRIMARY KEY,
	channel VARCHAR(20) NOT NULL UNIQUE,
	guild BIGINT NOT NULL,
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (guild) REFERENCES frontend_discord_guilds(scope) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS frontend_discord_users (
	scope BIGINT PRIMARY KEY,
	uid VARCHAR(20) NOT NULL UNIQUE,
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE
);

----------------------
--                  --
-- Frontend: Twitch --
--                  --
----------------------

CREATE TABLE IF NOT EXISTS frontend_twitch_channels (
	scope BIGINT PRIMARY KEY,
	channel_id VARCHAR(255) NOT NULL UNIQUE,
	channel_name VARCHAR(255) NOT NULL,
	access_token VARCHAR(255),
	refresh_token VARCHAR(255),
	FOREIGN KEY (scope) REFERENCES scopes(id) ON DELETE CASCADE
);

------------------------------
--                          --
-- Command: Custom Commands --
--                          --
------------------------------

CREATE TABLE IF NOT EXISTS cmd_customcommand_commands (
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

-------------------
--               --
-- Command: Time --
--               --
-------------------

CREATE TABLE IF NOT EXISTS cmd_time_reminders (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, 
	person BIGINT NOT NULL,
	place BIGINT NOT NULL,
	time INTEGER NOT NULL,
	what VARCHAR(255) NOT NULL,
	msg_id VARCHAR(255) NOT NULL,
	FOREIGN KEY (person) REFERENCES scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES scopes(id) ON DELETE CASCADE
);
