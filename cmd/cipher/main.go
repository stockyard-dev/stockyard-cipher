package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-cipher/internal/server";"github.com/stockyard-dev/stockyard-cipher/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="8870"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./cipher-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("cipher: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Cipher — Self-hosted password manager\n  ─────────────────────────────────\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Data:       %s\n  ─────────────────────────────────\n\n",port,port,dataDir)
log.Printf("cipher: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
