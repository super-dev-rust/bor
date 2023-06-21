package plaguedb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	_ "github.com/mattn/go-sqlite3"
)

func SaveTxs(txs []*types.Transaction, peerID string) {
	db, err := OpenDB()
	if err != nil {
		log.Warn("Failed to open the db:", "err", err)
	}
	defer db.Close()
	// for _, tx := range txs {
	// 	saveTx(db, tx, peerID)
	// 	log.Warn("Saving tx:", "tx", tx.Hash().Hex())
	// }
	for _, tx := range txs {
		if err := saveTx(db, tx, peerID); err != nil {
			log.Warn("Failed to save transaction:", "tx", tx.Hash().Hex(), "err", err)
		} else {
			log.Warn("Saved transaction:", "tx", tx.Hash().Hex())
		}
	}
}

func saveTx(db *sql.DB, tx *types.Transaction, peerID string) error {
	ts := time.Now().Unix()
	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	addr, err := signer.Sender(tx)
	if err != nil {
		log.Warn("Failed to get the sender:", "err", err)
		addr = common.HexToAddress("0x0")
	}
	//check if we have tx_hash in tx_summary
	var peerIDs string
	err = db.QueryRow("SELECT peer_id FROM tx_summary WHERE tx_hash = ?", tx.Hash().Hex()).Scan(&peerIDs)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("No rows found")
			if err := insertTxSummary(db, tx.Hash().Hex(), peerID, ts); err != nil {
				return err
			}
			// insertTxSummary(db, tx.Hash().Hex(), peerID, ts)
			// log.Warn("We're done with inserting tx summary", "db", db)
			//insert into txs
			// insertTxFetched(db, tx.Hash().Hex(), tx.GasPrice().Int64(), ts, addr.Hex(), tx.To().Hex())
			// log.Warn("We're done with inserting new rows")
			if err := insertTxFetched(db, tx.Hash().Hex(), tx.GasPrice().Int64(), ts, addr.Hex(), tx.To().Hex()); err != nil {
				return err
			}
			log.Info("Inserted newrows")
			return nil
		} else {
			log.Warn("Failed to query the tx:", "err", err)
		}
	}

	if !strings.Contains(peerIDs, peerID) {
		peerIDs = peerIDs + "," + peerID
		updateSQL := `UPDATE tx_summary SET peer_id = ? WHERE tx_hash = ?`
		_, err := db.Exec(updateSQL, peerIDs, tx.Hash().Hex())
		log.Warn("Updating tx_summary:", "peerIDS", peerIDs)
		if err != nil {
			log.Warn("Failed to update the tx:", "err", err)
			return err
		}
	}
	return nil
}

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "/Users/ako/bor/watcher.db")
	if err != nil {
		return nil, err
	}
	if err := prepareAndExecQuery(db, `CREATE TABLE IF NOT EXISTS tx_summary(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		tx_hash TEXT,
		peer_id TEXT,
		tx_first_seen INTEGER
	);`); err != nil {
		return nil, err
	}
	if err := prepareAndExecQuery(db, `CREATE TABLE IF NOT EXISTS tx_fetched(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		tx_hash TEXT,fee INTEGER,
		tx_first_seen INTEGER,
		"from" TEXT,
		"to" TEXT
		);`); err != nil {
		return nil, err
	}
	// initSQL = `CREATE TABLE IF NOT EXISTS tx_fetched(
	// 	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	// 	tx_hash TEXT,fee INTEGER,
	// 	tx_first_seen INTEGER,
	// 	"from" TEXT,
	// 	"to" TEXT
	// 	);`
	// prepareAndExecQuery(db, initSQL)

	return db, nil
}

func prepareAndExecQuery(db *sql.DB, queryString string) error {
	query, err := db.Prepare(queryString)
	if err != nil {
		return err
	}
	_, err = query.Exec()
	return err
}

func insertTxSummary(db *sql.DB, tx_hash string, peer_id string, tx_first_seen int64) error {
	insertSQL := `INSERT INTO tx_summary(tx_hash, peer_id, tx_first_seen) VALUES(?,?,?)`
	_, err := db.Exec(insertSQL, tx_hash, peer_id, tx_first_seen)
	return err
}

func insertTxFetched(db *sql.DB, tx_hash string, fee int64, tx_first_seen int64, from string, to string) error {
	insertSQL := `INSERT INTO tx_fetched(tx_hash, fee, tx_first_seen, "from", "to") VALUES(?,?,?,?,?)`
	_, err := db.Exec(insertSQL, tx_hash, fee, tx_first_seen, from, to)
	return err
}
