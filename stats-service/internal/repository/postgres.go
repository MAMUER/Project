package repository

import (
	"database/sql"
	"fmt"
	"stats-service/internal/model"
	"time"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(connStr string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Устанавливаем search_path для схемы
	_, err = db.Exec("SET search_path TO fitness_club_db")
	if err != nil {
		return nil, fmt.Errorf("failed to set search_path: %v", err)
	}

	return &PostgresRepo{db: db}, db.Ping()
}

func (r *PostgresRepo) GetTodayVisits(clubID int) (int, error) {
	var count int
	// Исправленный запрос с использованием связующей таблицы
	query := `
        SELECT COUNT(*) 
        FROM visits_history vh
        JOIN members_have_visits_history mvh ON vh.id_visit = mvh.id_visit
        JOIN members m ON m.id_member = mvh.id_member
        WHERE m.club_name = $1 AND vh.visit_date = CURRENT_DATE`

	err := r.db.QueryRow(query, clubID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresRepo) GetTopActiveMembers(limit int) ([]model.MemberActivity, error) {
	query := `
        SELECT 
            m.id_member, 
            m.first_name || ' ' || m.second_name as full_name,
            COUNT(vh.id_visit) as visits,
            MAX(vh.visit_date) as last_visit
        FROM members m
        JOIN members_have_visits_history mvh ON m.id_member = mvh.id_member
        JOIN visits_history vh ON vh.id_visit = mvh.id_visit
        WHERE vh.visit_date >= CURRENT_DATE - INTERVAL '30 days'
        GROUP BY m.id_member, m.first_name, m.second_name
        ORDER BY visits DESC
        LIMIT $1`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []model.MemberActivity
	for rows.Next() {
		var a model.MemberActivity
		var lastVisit time.Time
		err := rows.Scan(&a.MemberID, &a.MemberName, &a.VisitCount, &lastVisit)
		if err != nil {
			continue
		}
		a.LastVisit = lastVisit.Format("2006-01-02")
		activities = append(activities, a)
	}
	return activities, nil
}
