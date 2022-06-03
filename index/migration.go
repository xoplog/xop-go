



const setup = `
	CREATE TABLE IF NOT EXISTS attribues_names (
		name		text	PRIMARY KEY,
		anum		int	UNSINGED,
		type		int	UNSIGNED,

	CREATE TABLE IF NOT EXISTS interval_counts (
		start		timestamp with time zone,
		documents	bigint,
		bytes		bigint,
		attributes	bigint,
		blobs		bigint
	);

	CREATE TABLE IF NOT EXISTS document_json (
		id		uuid		NOT NULL,
		anum		int		UNSIGNED NOT NULL,
		blob		jsonb		NOT NULL,
		PRIMARY KEY (id, anum)
	);

	CREATE TABLE IF NOT EXISTS document_floats (
		id		uuid		NOT NULL,
		anum		int		UNSIGNED NOT NULL,
		number		float8		NOT NULL,
		PRIMARY KEY (id, anum)
	);

	CREATE TABLE IF NOT EXISTS document_integers (
		id		uuid		NOT NULL,
		anum		int		UNSIGNED NOT NULL,
		number		bigint		NOT NULL,
		PRIMARY KEY (id, anum)
	);

	CREATE TABLE IF NOT EXISTS document_strings (
		id		uuid		NOT NULL,
		anum		int		UNSIGNED NOT NULL,
		string		text		NOT NULL,	
		PRIMARY KEY (id, anum)
	);

	CREATE TABLE IF NOT EXISTS string_index (
		anum		int		UNSIGNED NOT NULL
		snippet		varchat(64)	NOT NULL,
		ts		datetime(6)	NOT NULL,
		docid		uuid		NOT NULL
		PRIMARY KEY (anum, snippet, ts, docid)
	);

	CREATE TABLE IF NOT EXISTS numeric_index (
		anum		int		UNSIGNED NOT NULL
		start		float8		NOT NULL,
		end		float8		NOT NULL,
		ts		datetime(6)	NOT NULL,
		docid		uuid		NOT NULL
		PRIMARY KEY (anum, start, end, ts, docid)
	);

	CREATE TABLE IF NOT EXISTS integer_index (
		anum		int		UNSIGNED NOT NULL,
		start		bigint		NOT NULL,
		end		bigint		NOT NULL,
		ts		datetime(6)	NOT NULL,
		docid		uuid		NOT NULL
		PRIMARY KEY (anum, start, end, ts, docid)
	);

	CREATE TABLE IF NOT EXISTS string_index_blobs (
		anum		int		UNSIGNED NOT NULL,
		snippet		varchar(64)	NOT NULL,


