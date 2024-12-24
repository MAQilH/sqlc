package admin

import (
	"errors"
	"fmt"
	"github.com/sqlc-dev/sqlc/internal/admin/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

type DatabaseRepository interface {
	GetTablesNames() (tablesName []string, err error)
	GetTableColumns(tableName string) (columns []model.ColumnInfo, err error)
	DropTable(tableName string) (err error)
	GetPrimaryColumn(tableName string) (columnName string, err error)
	CreateDocument(info model.DocumentInfo) (err error)
	GetTableDocuments(tableName string, pageNumber int, documentPerPage int) (documents []map[string]interface{}, err error)
	GetTableSize(tableName string) (tableSize int64, err error)
	EditDocument(tableName string, prevDocument map[string]string, updatedDocument map[string]string) (err error)
	DeleteDocument(tableName string, documentData map[string]string) (err error)
	FindAdminWithUsername(username string) (admin model.AdminSchema, err error)
	InsertAdmin(admin model.AdminSchema) (model.AdminSchema, error)
}

type DatabaseService struct {
	DB *gorm.DB
}

func NewDatabaseService() DatabaseRepository {
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")

	// connect to database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbName, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connection established")

	migrate(db)
	return &DatabaseService{db}
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(&model.AdminSchema{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database migrated!")
}

func (ds *DatabaseService) GetTableSize(tableName string) (tableSize int64, err error) {
	var count int64
	err = ds.DB.Table(tableName).Count(&count).Error
	return count, err
}

func (ds *DatabaseService) GetTablesNames() (tablesName []string, err error) {
	query := `SELECT tablename FROM pg_tables WHERE schemaname = 'public'`
	var tablesNames []string
	result := ds.DB.Raw(query).Scan(&tablesNames)
	if result.Error != nil {
		log.Fatalf("Error fetching table names: %v", result.Error)
		return nil, result.Error
	}
	return tablesNames, nil
}

func (ds *DatabaseService) GetTableColumns(tableName string) (columns []model.ColumnInfo, err error) {
	query := `
		SELECT column_name, column_default, is_nullable, data_type
		FROM information_schema.columns 
		WHERE table_name = ?;
	`
	result := ds.DB.Raw(query, tableName).Scan(&columns)
	return columns, result.Error
}

func (ds *DatabaseService) DropTable(tableName string) (err error) {
	return ds.DB.Migrator().DropTable(tableName)
}

func (ds *DatabaseService) GetPrimaryColumn(tableName string) (columnName string, err error) {
	query := `
		SELECT a.attname AS column_name
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = ?::regclass
		AND i.indisprimary;
	`
	var primaryColumnName string
	result := ds.DB.Raw(query, tableName).Scan(&primaryColumnName)
	return primaryColumnName, result.Error
}

func (ds *DatabaseService) CreateDocument(info model.DocumentInfo) (err error) {
	fmt.Println(info)

	columns := ""
	values := ""

	index := 0
	for columnName, columnValue := range info.FieldsInfo {
		if columnValue == "" {
			continue
		}

		if index > 0 {
			columns += ", "
			values += ", "
		}
		index++

		columns += columnName
		values += fmt.Sprintf("'%s'", columnValue)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", info.TableName, columns, values)

	if err = ds.DB.Exec(query).Error; err != nil {
		log.Printf("Failed to insert row: %v", err)
	}

	return err
}

func (ds *DatabaseService) GetTableDocuments(tableName string, pageNumber int, documentPerPage int) (documents []map[string]interface{}, err error) {
	err = ds.DB.Table(tableName).Limit(documentPerPage).Offset(pageNumber * documentPerPage).Find(&documents).Error
	if err != nil {
		log.Fatalf("Failed to query table: %v", err)
	}
	return documents, err
}

func (ds *DatabaseService) EditDocument(tableName string, prevDocument map[string]string, updatedDocument map[string]string) (err error) {
	setQuery := ""
	index := 0
	for columnName, columnValue := range updatedDocument {
		if index > 0 {
			setQuery += ", "
		}
		index++
		setQuery += fmt.Sprintf("%s = '%s'", columnName, columnValue)
	}

	index = 0
	conditionQuery := ""
	for columnName, columnValue := range prevDocument {
		if index > 0 {
			conditionQuery += " AND "
		}
		index++
		conditionQuery += fmt.Sprintf("%s = '%s'", columnName, columnValue)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", tableName, setQuery, conditionQuery)

	if err = ds.DB.Exec(query).Error; err != nil {
		log.Fatalf("Failed to query table: %v", err)
	}

	return err
}

func (ds *DatabaseService) DeleteDocument(tableName string, documentData map[string]string) (err error) {
	index := 0
	conditionQuery := ""
	for columnName, columnValue := range documentData {
		if index > 0 {
			conditionQuery += " AND "
		}
		index++
		conditionQuery += fmt.Sprintf("%s = '%s'", columnName, columnValue)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s;", tableName, conditionQuery)

	result := ds.DB.Exec(query)
	err = result.Error
	if result.RowsAffected == 0 {
		return errors.New("There isn't such document please refresh your table dashboard!")
	}
	return err
}

func (ds *DatabaseService) FindAdminWithUsername(username string) (admin model.AdminSchema, err error) {
	if err := ds.DB.First(&admin, "username = ?", username).Error; err != nil {
		log.Printf("Failed to query admin table: %v\n", err)
		return model.AdminSchema{}, err
	}
	return admin, err
}

func (ds *DatabaseService) InsertAdmin(admin model.AdminSchema) (model.AdminSchema, error) {
	err := ds.DB.Create(&admin).Error
	return admin, err
}
