package admin

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/sqlc-dev/sqlc/internal/admin/model"
	"github.com/sqlc-dev/sqlc/internal/admin/utils"
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
	DB *sql.DB
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

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connection established")

	//migrate(db)
	return &DatabaseService{db}
}

//func migrate(db *gorm.DB) {
//	err := db.AutoMigrate(&model.AdminSchema{})
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Database migrated!")
//}

func (ds *DatabaseService) GetTableSize(tableName string) (tableSize int64, err error) {
	var count int64
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("COUNT(*)").From(tableName)

	query, args, err := sb.ToSql()
	if err != nil {
		return 0, err
	}

	err = ds.DB.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (ds *DatabaseService) GetTablesNames() (tablesName []string, err error) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("table_name").
		From("information_schema.tables").
		Where(squirrel.Eq{"table_schema": "public"})

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := ds.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	for result.Next() {
		var table string
		if err := result.Scan(&table); err != nil {
			return nil, err
		}
		tablesName = append(tablesName, table)
	}

	fmt.Println("this is result", tablesName)
	return tablesName, nil
}

func (ds *DatabaseService) GetTableColumns(tableName string) (columns []model.ColumnInfo, err error) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("column_name", "column_default", "is_nullable", "data_type").
		From("information_schema.columns").
		Where(squirrel.Eq{"table_name": tableName})

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := ds.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	for result.Next() {
		var column model.ColumnInfo
		if err := result.Scan(&column.ColumnName, &column.ColumnDefault, &column.IsNullable, &column.DataType); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

func (ds *DatabaseService) DropTable(tableName string) (err error) {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err = ds.DB.Exec(query)

	return err
}

func (ds *DatabaseService) GetPrimaryColumn(tableName string) (columnName string, err error) {
	//query := `
	//	SELECT a.attname AS column_name
	//	FROM pg_index i
	//	JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
	//	WHERE i.indrelid = ?::regclass
	//	AND i.indisprimary;
	//`
	//var primaryColumnName string
	//result := ds.DB.Raw(query, tableName).Scan(&primaryColumnName)
	return "", nil
}

func (ds *DatabaseService) CreateDocument(info model.DocumentInfo) (err error) {
	fmt.Println(info)

	var columns []string
	var values []interface{}

	for columnName, columnValue := range info.FieldsInfo {
		columns = append(columns, columnName)
		values = append(values, columnValue)
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert(info.TableName).
		Columns(columns...).
		Values(values...)

	query, args, err := sb.ToSql()
	if err != nil {
		return err
	}

	_, err = ds.DB.Exec(query, args...)
	return err
}

func (ds *DatabaseService) GetTableDocuments(tableName string, pageNumber int, documentPerPage int) (documents []map[string]interface{}, err error) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("*").
		From(tableName).
		Limit(uint64(documentPerPage)).
		Offset(uint64(pageNumber * documentPerPage))

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := ds.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()

	for rows.Next() {
		var documentData = make([]interface{}, len(columns))
		var documentDataPtr = make([]interface{}, len(columns))
		for documentIndex := range len(columns) {
			documentDataPtr[documentIndex] = &documentData[documentIndex]
		}

		err = rows.Scan(documentDataPtr...)
		if err != nil {
			return nil, err
		}
		document := make(map[string]interface{})
		for index, data := range documentData {
			document[columns[index]] = data
		}
		documents = append(documents, document)
	}

	return documents, nil
}

func (ds *DatabaseService) EditDocument(tableName string, prevDocument map[string]string, updatedDocument map[string]string) (err error) {
	conditions := squirrel.And{}
	for key, value := range prevDocument {
		conditions = append(conditions, squirrel.Eq{key: value})
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Update(tableName).
		SetMap(utils.ConvertMapStringToMapInterface(updatedDocument)).
		Where(conditions)

	query, args, err := sb.ToSql()

	result, err := ds.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (ds *DatabaseService) DeleteDocument(tableName string, documentData map[string]string) (err error) {
	conditions := squirrel.And{}
	for key, value := range documentData {
		conditions = append(conditions, squirrel.Eq{key: value})
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Delete(tableName).
		Where(conditions)

	query, args, err := sb.ToSql()

	result, err := ds.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func (ds *DatabaseService) FindAdminWithUsername(username string) (admin model.AdminSchema, err error) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("*").From("admin_schemas").Where(squirrel.Eq{"username": username})

	query, args, err := sb.ToSql()
	if err != nil {
		return admin, err
	}

	err = ds.DB.QueryRow(query, args...).Scan(&admin.Username, &admin.Password, &admin.TelegramID, &admin.Email)
	if err != nil {
		return admin, err
	}
	return admin, nil
}

func (ds *DatabaseService) InsertAdmin(admin model.AdminSchema) (model.AdminSchema, error) {
	adminMapped := map[string]interface{}{
		"username":    admin.Username,
		"password":    admin.Password,
		"telegram_id": admin.TelegramID,
		"email":       admin.Email,
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("admin_schemas").
		SetMap(adminMapped)

	query, args, err := sb.ToSql()
	if err != nil {
		return admin, err
	}

	result, err := ds.DB.Exec(query, args...)
	if err != nil {
		return admin, err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return admin, err
	}

	return admin, nil
}
