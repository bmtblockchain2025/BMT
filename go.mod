module BMT-Blockchain

go 1.22

require (
    github.com/dgraph-io/badger/v3 v3.2103.2  // Lưu trữ blockchain
    github.com/gorilla/mux v1.8.0             // Router cho API
    github.com/lib/pq v1.10.0                 // Hỗ trợ PostgreSQL (nếu cần lưu blockchain vào DB)
    github.com/sirupsen/logrus v1.9.0         // Logging nâng cao
)
