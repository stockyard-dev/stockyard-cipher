package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Secret struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL string `json:"url"`
	Category string `json:"category"`
	Notes string `json:"notes"`
	Status string `json:"status"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"cipher.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS secrets(id TEXT PRIMARY KEY,name TEXT NOT NULL,username TEXT DEFAULT '',password TEXT DEFAULT '',url TEXT DEFAULT '',category TEXT DEFAULT '',notes TEXT DEFAULT '',status TEXT DEFAULT 'active',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Secret)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO secrets(id,name,username,password,url,category,notes,status,created_at)VALUES(?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.Username,e.Password,e.URL,e.Category,e.Notes,e.Status,e.CreatedAt);return err}
func(d *DB)Get(id string)*Secret{var e Secret;if d.db.QueryRow(`SELECT id,name,username,password,url,category,notes,status,created_at FROM secrets WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.Username,&e.Password,&e.URL,&e.Category,&e.Notes,&e.Status,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Secret{rows,_:=d.db.Query(`SELECT id,name,username,password,url,category,notes,status,created_at FROM secrets ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Secret;for rows.Next(){var e Secret;rows.Scan(&e.ID,&e.Name,&e.Username,&e.Password,&e.URL,&e.Category,&e.Notes,&e.Status,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *Secret)error{_,err:=d.db.Exec(`UPDATE secrets SET name=?,username=?,password=?,url=?,category=?,notes=?,status=? WHERE id=?`,e.Name,e.Username,e.Password,e.URL,e.Category,e.Notes,e.Status,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM secrets WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM secrets`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]Secret{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (name LIKE ?)"
        args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["category"];ok&&v!=""{where+=" AND category=?";args=append(args,v)}
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,name,username,password,url,category,notes,status,created_at FROM secrets WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []Secret;for rows.Next(){var e Secret;rows.Scan(&e.ID,&e.Name,&e.Username,&e.Password,&e.URL,&e.Category,&e.Notes,&e.Status,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM secrets GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
