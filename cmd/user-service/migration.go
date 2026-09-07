package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/MAMUER/project/internal/db"
	"github.com/MAMUER/project/internal/logger"
	"go.uber.org/zap"
)

// ensurePgsodiumKey идемпотентно импортирует PII-ключ из DB_ENCRYPTION_KEY
// в keyring pgsodium (таблица pgsodium.key) и фиксирует его идентификатор
// в пакете db для использования в шифровании/расшифровке.
func ensurePgsodiumKey(ctx context.Context, database *sql.DB, log *logger.Logger) error {
	key := db.EncryptionKey()
	if len(key) == 0 {
		return fmt.Errorf("DB_ENCRYPTION_KEY not set; pgsodium keyring cannot be initialized")
	}

	var id int64
	err := database.QueryRowContext(ctx, `SELECT id FROM pgsodium.key WHERE name = $1`, db.PgsodiumKeyringName()).Scan(&id)
	if err == nil {
		db.SetPgsodiumKeyID(id)
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("query pgsodium key: %w", err)
	}

	hexKey := hex.EncodeToString(key)
	err = database.QueryRowContext(ctx,
		`SELECT pgsodium.import_key(CASE WHEN $1 ~ '^[0-9a-fA-F]{64}$' THEN decode($1, 'hex') ELSE convert_to($1, 'UTF8') END, $2)`,
		hexKey, db.PgsodiumKeyringName(),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("import pgsodium key: %w", err)
	}
	db.SetPgsodiumKeyID(id)
	log.Info("Imported pgsodium PII key", zap.Int64("key_id", id))
	return nil
}

type pair struct {
	enc   string
	plain string
}

type piiTable struct {
	name  string
	idCol string
	pairs []pair
}

// reencryptPIIFromPgcrypto перекодирует существующие PII-поля,
// зашифрованные ранее через pgcrypto (pgp_sym_encrypt), в pgsodium (libsodium AEAD).
// Строки, уже зашифрованные через pgsodium, пропускаются.
func reencryptPIIFromPgcrypto(ctx context.Context, database *sql.DB, log *logger.Logger) {
	key := string(db.EncryptionKey())
	if key == "" {
		return
	}
	id := db.PgsodiumKeyID()
	if id == 0 {
		return
	}

	tables := []piiTable{
		{"users", "id", []pair{
			{"email_encrypted", "email"},
			{"full_name_encrypted", "full_name"},
			{"nickname_encrypted", "nickname"},
		}},
		{"email_verifications", "id", []pair{
			{"email_encrypted", "email"},
			{"token_encrypted", "token"},
		}},
	}

	for _, t := range tables {
		migrateTablePII(ctx, database, log, t, key, id)
	}
}

func migrateTablePII(ctx context.Context, database *sql.DB, log *logger.Logger, t piiTable, key string, id int64) {
	cols := []string{t.idCol}
	for _, p := range t.pairs {
		cols = append(cols, p.enc)
	}
	colList := strings.Join(cols, ", ")
	var selectBuilder strings.Builder
	selectBuilder.WriteString("SELECT ")
	selectBuilder.WriteString(colList)
	selectBuilder.WriteString(" FROM ")
	selectBuilder.WriteString(t.name)
	selectBuilder.WriteString(" WHERE ")
	selectBuilder.WriteString(t.pairs[0].enc)
	selectBuilder.WriteString(" IS NOT NULL")
	rows, err := database.QueryContext(ctx, selectBuilder.String())
	if err != nil {
		log.Error("Failed to scan PII rows for migration", zap.Error(err), zap.String("table", t.name))
		return
	}

	if err := rows.Err(); err != nil {
		log.Error("Failed to iterate PII rows for migration", zap.Error(err), zap.String("table", t.name))
		return
	}

	scanPtrs := make([]interface{}, len(cols))
	rowVals := make([]interface{}, len(cols))
	for i := range scanPtrs {
		scanPtrs[i] = &rowVals[i]
	}

	migrated := int64(0)
	for rows.Next() {
		if err := rows.Scan(scanPtrs...); err != nil {
			log.Error("Failed to scan PII row", zap.Error(err))
			continue
		}
		rowID := fmt.Sprint(rowVals[0])

		if migratePIIRow(ctx, database, log, t, key, id, rowID, rowVals) {
			migrated++
		}
	}
	if rowErr := rows.Err(); rowErr != nil {
		log.Error("Failed to iterate PII rows for migration", zap.Error(rowErr), zap.String("table", t.name))
		return
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Failed to close rows during PII migration", zap.Error(closeErr), zap.String("table", t.name))
	}
	if migrated > 0 {
		log.Info("Re-encrypted PII from pgcrypto to pgsodium", zap.String("table", t.name), zap.Int64("rows", migrated))
	}
}

func migratePIIRow(ctx context.Context, database *sql.DB, log *logger.Logger, t piiTable, key string, id int64, rowID string, rowVals []interface{}) bool {
	var probe string
	if database.QueryRowContext(ctx,
		fmt.Sprintf("SELECT convert_from(pgsodium.crypto_aead_det_decrypt($1, '', %d), 'UTF8')", id), rowVals[1],
	).Scan(&probe) == nil {
		return false
	}

	setParts := make([]string, 0, len(t.pairs))
	args := make([]interface{}, 0, len(t.pairs)+1)
	ai := 1
	for i, p := range t.pairs {
		var plain sql.NullString
		if dErr := database.QueryRowContext(ctx, "SELECT pgp_sym_decrypt($1, $2)", rowVals[i+1], key).Scan(&plain); dErr != nil || !plain.Valid {
			log.Warn("Failed to pgcrypto-decrypt during PII migration",
				zap.Error(dErr), zap.String("table", t.name), zap.String("col", p.enc))
			return false
		}
		args = append(args, plain.String)
		setParts = append(setParts, fmt.Sprintf("%s = pgsodium.crypto_aead_det_encrypt($%d::text, '', %d)", p.enc, ai, id))
		ai++
	}
	if len(setParts) == 0 {
		return false
	}
	args = append(args, rowID)

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE ")
	queryBuilder.WriteString(t.name)
	queryBuilder.WriteString(" SET ")
	queryBuilder.WriteString(strings.Join(setParts, ", "))
	queryBuilder.WriteString(" WHERE ")
	queryBuilder.WriteString(t.idCol)
	queryBuilder.WriteString(" = $")
	queryBuilder.WriteString(fmt.Sprintf("%d", ai))
	query := queryBuilder.String()

	if _, uErr := database.ExecContext(ctx, query, args...); uErr != nil {
		log.Error("Failed to re-encrypt PII row", zap.Error(uErr), zap.String("table", t.name), zap.String("id", rowID))
		return false
	}
	return true
}

// backfillEncryptedPII зашифровывает открытые PII-поля для существующих записей,
// у которых ещё нет pgsodium-шифротекста.
func backfillEncryptedPII(ctx context.Context, database *sql.DB, log *logger.Logger) {
	key := string(db.EncryptionKey())
	if key == "" {
		log.Warn("DB_ENCRYPTION_KEY not set; skipping PII backfill")
		return
	}
	id := db.PgsodiumKeyID()
	if id == 0 {
		log.Warn("pgsodium key not initialized; skipping PII backfill")
		return
	}
	enc := func(col string) string {
		return fmt.Sprintf("pgsodium.crypto_aead_det_encrypt(%s::text, '', %d)", col, id)
	}

	var usersQuery strings.Builder
	usersQuery.WriteString("UPDATE users SET ")
	usersQuery.WriteString("email_encrypted = ")
	usersQuery.WriteString(enc("email"))
	usersQuery.WriteString(", email_hash = encode(digest(lower(email), 'sha256'), 'hex'), ")
	usersQuery.WriteString("full_name_encrypted = ")
	usersQuery.WriteString(enc("full_name"))
	usersQuery.WriteString(", full_name_hash = encode(digest(lower(full_name), 'sha256'), 'hex'), ")
	usersQuery.WriteString("nickname_encrypted = ")
	usersQuery.WriteString(enc("nickname"))
	usersQuery.WriteString(", nickname_hash = encode(digest(lower(nickname), 'sha256'), 'hex') ")
	usersQuery.WriteString(" WHERE email_encrypted IS NULL")
	res, err := database.ExecContext(ctx, usersQuery.String())
	if err != nil {
		log.Error("Failed to backfill PII in users", zap.Error(err))
	} else {
		rows, _ := res.RowsAffected()
		log.Info("PII backfill complete for users", zap.Int64("updated", rows))
	}

	var emailVerificationsQuery strings.Builder
	emailVerificationsQuery.WriteString("UPDATE email_verifications SET ")
	emailVerificationsQuery.WriteString("email_encrypted = ")
	emailVerificationsQuery.WriteString(enc("email"))
	emailVerificationsQuery.WriteString(", token_encrypted = ")
	emailVerificationsQuery.WriteString(enc("token"))
	emailVerificationsQuery.WriteString(" WHERE email_encrypted IS NULL")
	_, err = database.ExecContext(ctx, emailVerificationsQuery.String())
	if err != nil {
		log.Error("Failed to backfill PII in email_verifications", zap.Error(err))
	}
}
