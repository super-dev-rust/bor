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

func SaveTxs(db *sql.DB, txs []*types.Transaction, peerID string) error {
	for _, tx := range txs {
		if err := saveTx(db, tx, peerID); err != nil {
			log.Warn("Failed to save transaction:", "tx", tx.Hash().Hex(), "err", err)
			return err
		} else {
			log.Warn("Saved transaction:", "tx", tx.Hash().Hex())
		}
	}
	return nil
}

func saveTx(db *sql.DB, tx *types.Transaction, peerID string) error {
	ts := time.Now().Unix()
	log.Warn("Saving tx:", "tx", tx)
	if tx.Protected() {
		log.Warn("Skipping protected tx:", "tx", tx.Hash().Hex())
		// return nil
	}
	//new
	signer := types.NewEIP2930Signer(tx.ChainId())

	addr, err := types.Sender(signer, tx)

	if err != nil {
		log.Warn("Failed to get the sender:", "err", err)
		addr = common.HexToAddress("0x438308")
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
			var to string
			if tx.To() == nil {
				to = "0x0"
			} else {
				to = tx.To().Hex()
			}

			log.Warn("Between summary und fetched")
			log.Warn("Each arg one by one")
			log.Warn("tx.Hash().Hex():", "Hex", tx.Hash().Hex())
			log.Warn("tx.GasPrice().Int64():", "int64", tx.GasPrice().Int64())
			log.Warn("ts:", "ts", ts)
			log.Warn("addr.Hex():", "Addr.Hex", addr.Hex())
			log.Warn("tx.To().Hex():", "tx.to()", to)

			if err := insertTxFetched(db, tx.Hash().Hex(), tx.GasPrice().Int64(), ts, addr.Hex(), to); err != nil {
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

func OpenDB(path string) (*sql.DB, error) {
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
	log.Warn("I smell it's there one", "tx_hash", tx_hash, "peer_id", peer_id, "tx_first_seen", tx_first_seen)
	insertSQL := `INSERT INTO tx_summary(tx_hash, peer_id, tx_first_seen) VALUES(?,?,?)`
	log.Warn("InsertSQL", "insertSQL", insertSQL)
	_, err := db.Exec(insertSQL, tx_hash, peer_id, tx_first_seen)
	log.Warn("After exec", "err", err)
	return err
}

func insertTxFetched(db *sql.DB, tx_hash string, fee int64, tx_first_seen int64, from string, to string) error {
	log.Warn("I smell it's there two")
	insertSQL := `INSERT INTO tx_fetched(tx_hash, fee, tx_first_seen, "from", "to") VALUES(?,?,?,?,?)`
	_, err := db.Exec(insertSQL, tx_hash, fee, tx_first_seen, from, to)
	return err
}
