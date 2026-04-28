package tests

import (
	"os"
	"testing"
	"time"

	"final/pkg/db"  
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// Task — локальная модель для тестов (совпадает с pkg/db.Task, но для sqlx)
type Task struct {
	ID      int64  `db:"id"`
	Date    string `db:"date"`
	Title   string `db:"title"`
	Comment string `db:"comment"`
	Repeat  string `db:"repeat"`
}

// count возвращает количество задач в БД
func count(db *sqlx.DB) (int, error) {
	var count int
	return count, db.Get(&count, `SELECT count(id) FROM scheduler`)
}

// openDB подключается к БД и гарантирует, что схема создана
func openDB(t *testing.T) *sqlx.DB {
	dbfile := DBFile
	envFile := os.Getenv("TODO_DBFILE")
	if len(envFile) > 0 {
		dbfile = envFile
	}

	// Инициализация схемы
	_, err := db.NewStore(dbfile)
	assert.NoError(t, err, "Не удалось инициализировать схему БД")

	// Подключение через sqlx
	dbConn, err := sqlx.Connect("sqlite", dbfile+"?_loc=auto")
	assert.NoError(t, err)

	//  Очистка таблицы перед тестом (изоляция!)
	_, err = dbConn.Exec(`DELETE FROM scheduler`)
	assert.NoError(t, err, "Не удалось очистить таблицу")

	//  Сброс автоинкремента ID (чтобы ID начинались с 1)
	_, err = dbConn.Exec(`DELETE FROM sqlite_sequence WHERE name='scheduler'`)
	// Игнорируем ошибку, если таблица sqlite_sequence ещё пуста

	return dbConn
}

// TestDB проверяет базовые операции с БД
func TestDB(t *testing.T) {
	dbConn := openDB(t)
	defer dbConn.Close()

	before, err := count(dbConn)
	assert.NoError(t, err)

	today := time.Now().Format(`20060102`)

	res, err := dbConn.Exec(`INSERT INTO scheduler (date, title, comment, repeat) 
	VALUES (?, 'Todo', 'Комментарий', '')`, today)
	assert.NoError(t, err)

	id, err := res.LastInsertId()
	assert.NoError(t, err)

	var task Task
	err = dbConn.Get(&task, `SELECT * FROM scheduler WHERE id=?`, id)
	assert.NoError(t, err)
	assert.Equal(t, id, task.ID)
	assert.Equal(t, `Todo`, task.Title)
	assert.Equal(t, `Комментарий`, task.Comment)

	_, err = dbConn.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	assert.NoError(t, err)

	after, err := count(dbConn)
	assert.NoError(t, err)

	assert.Equal(t, before, after)
}