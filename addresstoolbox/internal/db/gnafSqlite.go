package gnafSqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// GnafRecord represents a single GNAF core record
type GnafRecord struct {
	AddressDetailPID string
	DateCreated      string
	AddressLabel     string
	AddressSiteName  string
	BuildingName     string
	FlatType         string
	FlatNumber       string
	LevelType        string
	LevelNumber      string
	NumberFirst      string
	NumberLast       string
	LotNumber        string
	StreetName       string
	StreetType       string
	StreetSuffix     string
	LocalityName     string
	State            string
	Postcode         string
	LegalParcelID    string
	MBCode           string
	AliasPrincipal   string
	PrincipalPID     string
	PrimarySecondary string
	PrimaryPID       string
	GeocodeType      string
	Longitude        float64
	Latitude         float64
}

// GnafDB wraps the SQLite database connection
type GnafDB struct {
	db *sql.DB
}

// NewGnafDB initializes a new SQLite database connection
func NewGnafDB(dbPath string) (*GnafDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &GnafDB{db: db}, nil
}

// Close closes the database connection
func (g *GnafDB) Close() error {
	return g.db.Close()
}

// OptimizeForBulkInsert applies PRAGMA settings for faster bulk inserts
func (g *GnafDB) OptimizeForBulkInsert() error {
	pragmas := []string{
		"PRAGMA synchronous = OFF",
		"PRAGMA journal_mode = MEMORY",
		"PRAGMA cache_size = -200000", //200Mb
		"PRAGMA locking_mode = EXCLUSIVE",
	}
	for _, pragma := range pragmas {
		_, err := g.db.Exec(pragma)
		if err != nil {
			return fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}
	return nil
}

// OptimizeForReadPerformancePragmas applies PRAGMA settings for faster reads
func (g *GnafDB) OptimizeForReadPerformancePragmas() error {
	pragmas := []string{
		"PRAGMA cache_size = -200000", //200Mb
		"PRAGMA synchronous = NORMAL",
		"PRAGMA journal_mode = WAL",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = -200000", //200Mb
	}
	for _, pragma := range pragmas {
		_, err := g.db.Exec(pragma)
		if err != nil {
			return fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}
	return nil
}

// ResetPragmas restores safer default PRAGMA settings
func (g *GnafDB) ResetPragmas() error {
	pragmas := []string{
		"PRAGMA synchronous = FULL",
		"PRAGMA journal_mode = DELETE",
		"PRAGMA locking_mode = NORMAL",
	}
	for _, pragma := range pragmas {
		_, err := g.db.Exec(pragma)
		if err != nil {
			return fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}
	return nil
}

// Drop GNAF Table if exists
func (g *GnafDB) DropGnafTable(tableName string) error {
	//check if tableName is empty
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	dropTableSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)

	_, err := g.db.Exec(dropTableSQL)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}
	return nil
}

// CreateGnafTable creates the GNAF core table if it doesn't exist
func (g *GnafDB) CreateGnafTable(tableName string) error {
	//check if tableName is empty
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		ADDRESS_DETAIL_PID TEXT PRIMARY KEY,
		DATE_CREATED TEXT,
		ADDRESS_LABEL TEXT,
		ADDRESS_SITE_NAME TEXT,
		BUILDING_NAME TEXT,
		FLAT_TYPE TEXT,
		FLAT_NUMBER TEXT,
		LEVEL_TYPE TEXT,
		LEVEL_NUMBER TEXT,
		NUMBER_FIRST TEXT,
		NUMBER_LAST TEXT,
		LOT_NUMBER TEXT,
		STREET_NAME TEXT,
		STREET_TYPE TEXT,
		STREET_SUFFIX TEXT,
		LOCALITY_NAME TEXT,
		STATE TEXT,
		POSTCODE TEXT,
		LEGAL_PARCEL_ID TEXT,
		MB_CODE TEXT,
		ALIAS_PRINCIPAL TEXT,
		PRINCIPAL_PID TEXT,
		PRIMARY_SECONDARY TEXT,
		PRIMARY_PID TEXT,
		GEOCODE_TYPE TEXT,
		LONGITUDE REAL,
		LATITUDE REAL
	);`, tableName)

	_, err := g.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// InsertGnafRecordsBatch inserts multiple GNAF records in a single transaction
func (g *GnafDB) InsertGnafRecordsBatch(records []GnafRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := g.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
	INSERT INTO gnaf_core (
		ADDRESS_DETAIL_PID, DATE_CREATED, ADDRESS_LABEL, ADDRESS_SITE_NAME, BUILDING_NAME,
		FLAT_TYPE, FLAT_NUMBER, LEVEL_TYPE, LEVEL_NUMBER, NUMBER_FIRST, NUMBER_LAST, LOT_NUMBER,
		STREET_NAME, STREET_TYPE, STREET_SUFFIX, LOCALITY_NAME, STATE, POSTCODE,
		LEGAL_PARCEL_ID, MB_CODE, ALIAS_PRINCIPAL, PRINCIPAL_PID, PRIMARY_SECONDARY,
		PRIMARY_PID, GEOCODE_TYPE, LONGITUDE, LATITUDE
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		_, err := stmt.Exec(
			record.AddressDetailPID, record.DateCreated, record.AddressLabel, record.AddressSiteName, record.BuildingName,
			record.FlatType, record.FlatNumber, record.LevelType, record.LevelNumber, record.NumberFirst, record.NumberLast, record.LotNumber,
			record.StreetName, record.StreetType, record.StreetSuffix, record.LocalityName, record.State, record.Postcode,
			record.LegalParcelID, record.MBCode, record.AliasPrincipal, record.PrincipalPID, record.PrimarySecondary,
			record.PrimaryPID, record.GeocodeType, record.Longitude, record.Latitude,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// GetGnafRecordByPID retrieves a GNAF record by ADDRESS_DETAIL_PID
func (g *GnafDB) GetGnafRecordByPID(pid string) (*GnafRecord, error) {
	querySQL := `SELECT * FROM gnaf_core WHERE ADDRESS_DETAIL_PID = ?`
	row := g.db.QueryRow(querySQL, pid)

	var record GnafRecord
	err := row.Scan(
		&record.AddressDetailPID, &record.DateCreated, &record.AddressLabel, &record.AddressSiteName, &record.BuildingName,
		&record.FlatType, &record.FlatNumber, &record.LevelType, &record.LevelNumber, &record.NumberFirst, &record.NumberLast, &record.LotNumber,
		&record.StreetName, &record.StreetType, &record.StreetSuffix, &record.LocalityName, &record.State, &record.Postcode,
		&record.LegalParcelID, &record.MBCode, &record.AliasPrincipal, &record.PrincipalPID, &record.PrimarySecondary,
		&record.PrimaryPID, &record.GeocodeType, &record.Longitude, &record.Latitude,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no record found with ADDRESS_DETAIL_PID: %s", pid)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query record: %w", err)
	}
	return &record, nil
}
