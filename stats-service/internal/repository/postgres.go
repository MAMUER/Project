package repository

import (
    "database/sql"
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
    return &PostgresRepo{db: db}, db.Ping()
}

func (r *PostgresRepo) GetTodayVisits(clubID int) (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM visits_history 
              WHERE club_id = $1 AND DATE(visit_date) = CURRENT_DATE`
    err := r.db.QueryRow(query, clubID).Scan(&count)
    return count, err
}

func (r *PostgresRepo) GetTopActiveMembers(limit int) ([]model.MemberActivity, error) {
    rows, err := r.db.Query(`
        SELECT m.id, m.name, COUNT(vh.id) as visits, 
               MAX(vh.visit_date) as last_visit
        FROM members m
        JOIN visits_history vh ON m.id = vh.member_id
        WHERE vh.visit_date >= NOW() - INTERVAL '30 days'
        GROUP BY m.id, m.name
        ORDER BY visits DESC
        LIMIT $1`, limit)
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