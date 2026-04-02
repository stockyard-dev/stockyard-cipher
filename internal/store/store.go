package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Entry struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL string `json:"url"`
	Category string `json:"category"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"cipher.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS entries(id TEXT PRIMARY KEY,name TEXT NOT NULL,username TEXT DEFAULT '',password TEXT DEFAULT '',url TEXT DEFAULT '',category TEXT DEFAULT '',notes TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Entry)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO entries(id,name,username,password,url,category,notes,created_at)VALUES(?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.Username,e.Password,e.URL,e.Category,e.Notes,e.CreatedAt);return err}
func(d *DB)Get(id string)*Entry{var e Entry;if d.db.QueryRow(`SELECT id,name,username,password,url,category,notes,created_at FROM entries WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.Username,&e.Password,&e.URL,&e.Category,&e.Notes,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Entry{rows,_:=d.db.Query(`SELECT id,name,username,password,url,category,notes,created_at FROM entries ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Entry;for rows.Next(){var e Entry;rows.Scan(&e.ID,&e.Name,&e.Username,&e.Password,&e.URL,&e.Category,&e.Notes,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM entries WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&n);return n}
